/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package healthz

import "time"

// SetNow is a test function to set now
func (h *HealthHandler) SetNow(now func() time.Time) {
	h.now = now
}

// HealthCheckers is a test function which returns the map of HealthCheckers
func (h *HealthHandler) HealthCheckers() map[string]HealthChecker {
	return h.healthCheckers
}
