/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package healthz_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/hyperledger/fabric-lib-go/healthz"
	"github.com/hyperledger/fabric-lib-go/healthz/mock"
	. "github.com/onsi/gomega"
)

func TestRegisterChecker(t *testing.T) {
	t.Parallel()
	g := NewGomegaWithT(t)

	handler := healthz.NewHealthHandler()
	checker1 := &mock.HealthChecker{}
	checker2 := &mock.HealthChecker{}
	component1 := "component1"
	component2 := "component2"

	err := handler.RegisterChecker(component1, checker1)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(handler.HealthCheckers()).To(HaveKey(component1))

	err = handler.RegisterChecker(component1, checker1)
	g.Expect(err).To(MatchError(healthz.AlreadyRegisteredError(component1)))

	err = handler.RegisterChecker(component2, checker2)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(handler.HealthCheckers()).To(HaveLen(2))
	g.Expect(handler.HealthCheckers()).To(HaveKey(component1))
	g.Expect(handler.HealthCheckers()).To(HaveKey(component2))

	handler.DeregisterChecker(component1)
	g.Expect(handler.HealthCheckers()).To(HaveLen(1))
	g.Expect(handler.HealthCheckers()).ToNot(HaveKey(component1))
	g.Expect(handler.HealthCheckers()).To(HaveKey(component2))

	// deregister non-existent checker should be a no-op
	handler.DeregisterChecker(component1)
	g.Expect(handler.HealthCheckers()).To(HaveLen(1))
}

func TestRunChecks(t *testing.T) {
	t.Parallel()
	g := NewGomegaWithT(t)

	handler := healthz.NewHealthHandler()
	ctx := context.Background()

	checker := &mock.HealthChecker{}
	checker.HealthCheckReturns(nil)
	handler.RegisterChecker("good_component", checker)
	fc := handler.RunChecks(ctx)
	g.Expect(fc).To(HaveLen(0))

	failedChecker := &mock.HealthChecker{}
	reason := "poorly written code"
	failedChecker.HealthCheckReturnsOnCall(0, errors.New(reason))
	failedChecker.HealthCheckReturnsOnCall(1, errors.New(reason))
	handler.RegisterChecker("bad_component1", failedChecker)
	handler.RegisterChecker("bad_component2", failedChecker)
	fc = handler.RunChecks(ctx)
	g.Expect(failedChecker.HealthCheckCallCount()).To(Equal(2))
	g.Expect(fc).To(HaveLen(2))
	g.Expect(fc).To(ContainElement(
		healthz.FailedCheck{
			Component: "bad_component1",
			Reason:    reason,
		},
	))
	g.Expect(fc).To(ContainElement(
		healthz.FailedCheck{
			Component: "bad_component2",
			Reason:    reason,
		},
	))
}

func TestServeHTTP(t *testing.T) {
	t.Parallel()

	var expectedTime = time.Now()
	var tests = []struct {
		name           string
		healthCheckers map[string]healthz.HealthChecker
		failedChecks   []healthz.FailedCheck
		expectedCode   int
		expectedStatus string
	}{
		{
			name: "Status OK",
			healthCheckers: map[string]healthz.HealthChecker{
				"component1": &mock.HealthChecker{
					HealthCheckStub: func(context.Context) error {
						return nil
					},
				},
			},
			expectedCode:   http.StatusOK,
			expectedStatus: healthz.StatusOK,
		},
		{
			name: "Service Unavailable",
			healthCheckers: map[string]healthz.HealthChecker{
				"component1": &mock.HealthChecker{
					HealthCheckStub: func(context.Context) error {
						return errors.New("poorly written code")
					},
				},
			},
			failedChecks: []healthz.FailedCheck{
				{
					Component: "component1",
					Reason:    "poorly written code",
				},
			},
			expectedCode:   http.StatusServiceUnavailable,
			expectedStatus: healthz.StatusUnavailable,
		},
		{
			name: "Service Unavailable - Multiple",
			healthCheckers: map[string]healthz.HealthChecker{
				"component1": &mock.HealthChecker{
					HealthCheckStub: func(context.Context) error {
						return errors.New("poorly written code")
					},
				},
				"component2": &mock.HealthChecker{
					HealthCheckStub: func(context.Context) error {
						return errors.New("more poorly written code")
					},
				},
			},
			failedChecks: []healthz.FailedCheck{
				{
					Component: "component1",
					Reason:    "poorly written code",
				},
				{
					Component: "component2",
					Reason:    "more poorly written code",
				},
			},
			expectedCode:   http.StatusServiceUnavailable,
			expectedStatus: healthz.StatusUnavailable,
		},
		{
			name: "Mixed",
			healthCheckers: map[string]healthz.HealthChecker{
				"component1": &mock.HealthChecker{
					HealthCheckStub: func(context.Context) error {
						return errors.New("poorly written code")
					},
				},
				"component2": &mock.HealthChecker{
					HealthCheckStub: func(context.Context) error {
						return nil
					},
				},
			},
			failedChecks: []healthz.FailedCheck{
				{
					Component: "component1",
					Reason:    "poorly written code",
				},
			},
			expectedCode:   http.StatusServiceUnavailable,
			expectedStatus: healthz.StatusUnavailable,
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			g := NewGomegaWithT(t)
			hh := healthz.NewHealthHandler()
			hh.SetNow(func() time.Time { return expectedTime })
			for name, checker := range test.healthCheckers {
				err := hh.RegisterChecker(name, checker)
				if err != nil {
					t.Fatalf("Failed to register checker [%s]", err)
				}
			}
			ts := httptest.NewServer(hh)
			defer ts.Close()

			res, err := http.Get(ts.URL)
			if err != nil {
				t.Fatalf("Error getting response [%s]", err)
			}
			g.Expect(res.StatusCode).To(Equal(test.expectedCode))
			body, err := ioutil.ReadAll(res.Body)
			res.Body.Close()
			if err != nil {
				t.Fatalf("Error reading response body [%s]", err)
			}
			var hs healthz.HealthStatus
			err = json.Unmarshal(body, &hs)
			if err != nil {
				t.Fatalf("Error unmarshaling response body [%s]", err)
			}
			g.Expect(hs.Status).To(Equal(test.expectedStatus))
			g.Expect(hs.Time).To(BeTemporally("==", expectedTime))
			for _, failedCheck := range test.failedChecks {
				g.Expect(hs.FailedChecks).To(ContainElement(failedCheck))
			}
		})
	}
}

func TestServeHTTP_Timeout(t *testing.T) {
	t.Parallel()
	g := NewGomegaWithT(t)

	hc := &mock.HealthChecker{
		HealthCheckStub: func(ctx context.Context) error {
			select {
			case <-ctx.Done():
				t.Log("done called")
				return errors.New("check failed")
			}
		},
	}

	hh := healthz.NewHealthHandler()
	hh.SetTimeout(time.Second)
	err := hh.RegisterChecker("long_running", hc)
	if err != nil {
		t.Fatalf("error registering checker [%s]", err)
	}
	ts := httptest.NewServer(hh)
	defer ts.Close()

	res, err := ts.Client().Get(ts.URL)
	if err != nil {
		t.Fatalf("Error getting response [%s]", err)
	}
	g.Expect(res.StatusCode).To(Equal(http.StatusRequestTimeout))

	req, err := http.NewRequest(http.MethodGet, ts.URL, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("Failed to create request [%s]", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()
	res, err = ts.Client().Do(req.WithContext(ctx))
	g.Expect(res).To(BeNil())
	g.Expect(err).To(HaveOccurred())

}

func TestHandleGetOnly(t *testing.T) {
	t.Parallel()
	g := NewGomegaWithT(t)

	hh := healthz.NewHealthHandler()
	ts := httptest.NewServer(hh)
	defer ts.Close()

	methods := []string{
		http.MethodConnect,
		http.MethodDelete,
		http.MethodHead,
		http.MethodOptions,
		http.MethodPost,
		http.MethodPut,
		http.MethodTrace,
	}
	for _, method := range methods {
		req, err := http.NewRequest(method, ts.URL, &bytes.Buffer{})
		if err != nil {
			t.Fatalf("Failed to create request [%s]", err)
		}
		res, err := ts.Client().Do(req)
		g.Expect(err).NotTo(HaveOccurred())
		g.Expect(res.StatusCode).To(Equal(http.StatusMethodNotAllowed))
	}
}
