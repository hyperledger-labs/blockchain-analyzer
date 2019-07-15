package templates

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"

	"github.com/balazsprehoda/fabricbeat/modules/fabricbeatsetup"
	"github.com/elastic/beats/libbeat/logp"
)

// Generates the index patterns and dashboards for the connected peer from templates in the kibana_templates folder.
func GenerateDashboards(setup *fabricbeatsetup.FabricbeatSetup) error {

	// The beginnings of the dashboard template names (i.e. overview-dashboard-TEMPLATE.json -> overview)
	dashboardNames := []string{"overview", "block", "key", "transaction"}
	visualizationNames := []string{"block_count", "transaction_count", "transaction_per_organization", "transaction_count_timeline"}
	templates := []string{"block", "transaction", "key"}
	var patternId string
	// Create index patterns for the peer the agent connects to
	for _, templateName := range templates {
		// Load index pattern template and replace title
		logp.Info("Creating %s index pattern for connected peer", templateName)

		indexPatternJSON, err := ioutil.ReadFile(fmt.Sprintf("%s/%s-index-pattern-TEMPLATE.json", setup.TemplateDirectory, templateName))
		if err != nil {
			return err
		}

		indexPatternJSONstring := string(indexPatternJSON)
		// Replace id placeholders (in URL formatted fields)
		for _, dashboardName := range dashboardNames {

			// Replace dashboard id placeholders
			idExpression := fmt.Sprintf("%s_DASHBOARD_TEMPLATE_ID", strings.ToUpper(dashboardName))
			re := regexp.MustCompile(idExpression)
			indexPatternJSONstring = re.ReplaceAllString(indexPatternJSONstring, fmt.Sprintf("%s-dashboard-%s-%s", dashboardName, setup.Peer, setup.OrgName))

			// Replace search id placeholders
			idExpression = fmt.Sprintf("%s_SEARCH_TEMPLATE_ID", strings.ToUpper(dashboardName))
			re = regexp.MustCompile(idExpression)
			indexPatternJSONstring = re.ReplaceAllString(indexPatternJSONstring, fmt.Sprintf("%s-search-%s-%s", dashboardName, setup.Peer, setup.OrgName))

			// Replace visualization id placeholders
			idExpression = fmt.Sprintf("%s_VISUALIZATION_TEMPLATE_ID", strings.ToUpper(dashboardName))
			re = regexp.MustCompile(idExpression)
			indexPatternJSONstring = re.ReplaceAllString(indexPatternJSONstring, fmt.Sprintf("%s-visualization-%s-%s", dashboardName, setup.Peer, setup.OrgName))
		}

		// Replace title placeholders
		titleExpression := fmt.Sprintf("INDEX_PATTERN_TEMPLATE_TITLE")
		re := regexp.MustCompile(titleExpression)
		indexPatternJSONstring = re.ReplaceAllString(indexPatternJSONstring, fmt.Sprintf("fabricbeat*%s*%s", templateName, setup.OrgName))

		indexPatternJSON = []byte(indexPatternJSONstring)
		// Send index pattern to Kibana via Kibana Saved Objects API
		logp.Info("Persisting %s index pattern for connected peer", templateName)

		patternId = fmt.Sprintf("fabricbeat-%s-%s", templateName, setup.Peer)
		request, err := http.NewRequest("POST", fmt.Sprintf("%s/api/saved_objects/index-pattern/%s", setup.KibanaURL, patternId), bytes.NewBuffer(indexPatternJSON))
		if err != nil {
			return err
		}
		request.Header.Add("Content-Type", "application/json")
		request.Header.Add("kbn-xsrf", "true")
		httpClient := http.Client{}
		resp, err := httpClient.Do(request)
		if err != nil {
			return err
		}
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		// Check if the index pattern exists. 409 is for version conflict, it means that this index pattern has already been created.
		// TODO check for existing index pattern and replace it.
		if resp.StatusCode != 200 && resp.StatusCode != 409 {
			return errors.New(fmt.Sprintf("Failed to create %s index pattern:\nResponse status code: %d\nResponse body: %s", templateName, resp.StatusCode, string(body)))
		}
		logp.Info("%s index pattern created", templateName)
	}

	for _, dashboardName := range dashboardNames {

		// Load dashboard template
		logp.Info("Creating %s dashboard from template", dashboardName)
		dashboardBytes, err := ioutil.ReadFile(fmt.Sprintf("/home/prehi/go/src/github.com/balazsprehoda/fabricbeat/kibana_templates/%s-dashboard-TEMPLATE.json", dashboardName))
		if err != nil {
			return err
		}
		dashboard := string(dashboardBytes)

		for _, templateName := range templates {
			// Replace index pattern id placeholders
			patternId = fmt.Sprintf("fabricbeat-%s-%s", templateName, setup.Peer)
			patternExpression := fmt.Sprintf("%s_PATTERN", strings.ToUpper(templateName))
			re := regexp.MustCompile(patternExpression)
			dashboard = re.ReplaceAllString(string(dashboard), patternId)

			// Replace search id placeholders
			searchId := fmt.Sprintf("%s-search-%s-%s", templateName, setup.Peer, setup.OrgName)
			searchIdExpression := fmt.Sprintf("%s_SEARCH_TEMPLATE_ID", strings.ToUpper(templateName))
			re = regexp.MustCompile(searchIdExpression)
			dashboard = re.ReplaceAllString(dashboard, searchId)

			// Replace search title placeholders
			searchTitle := fmt.Sprintf("%s Search %s (%s)", strings.Title(templateName), setup.Peer, setup.OrgName)
			searchTitleExpression := fmt.Sprintf("%s_SEARCH_TEMPLATE_TITLE", strings.ToUpper(templateName))
			re = regexp.MustCompile(searchTitleExpression)
			dashboard = re.ReplaceAllString(dashboard, searchTitle)
		}

		for _, visualizationName := range visualizationNames {
			// Replace visualization id placeholders
			visualizationId := fmt.Sprintf("%s-visualization-%s-%s", visualizationName, setup.Peer, setup.OrgName)
			visualizationIdExpression := fmt.Sprintf("%s_VISUALIZATION_TEMPLATE_ID", strings.ToUpper(visualizationName))
			re := regexp.MustCompile(visualizationIdExpression)
			dashboard = re.ReplaceAllString(dashboard, visualizationId)

			// Replace visualization title placeholders
			visualizationTitle := fmt.Sprintf("%s Visualization %s (%s)", strings.Title(visualizationName), setup.Peer, setup.OrgName)
			visualizationTitleExpression := fmt.Sprintf("%s_VISUALIZATION_TEMPLATE_TITLE", strings.ToUpper(visualizationName))
			re = regexp.MustCompile(visualizationTitleExpression)
			dashboard = re.ReplaceAllString(dashboard, visualizationTitle)
		}

		// Replace dashboard id
		idExpression := fmt.Sprintf("%s_DASHBOARD_TEMPLATE_ID", strings.ToUpper(dashboardName))
		re := regexp.MustCompile(idExpression)
		dashboard = re.ReplaceAllString(string(dashboard), fmt.Sprintf("%s-dashboard-%s-%s", dashboardName, setup.Peer, setup.OrgName))

		// Replace dashboard title
		titleExpression := fmt.Sprintf("%s_DASHBOARD_TEMPLATE_TITLE", strings.ToUpper(dashboardName))
		re = regexp.MustCompile(titleExpression)
		dashboard = re.ReplaceAllString(string(dashboard), fmt.Sprintf("%s Dashboard %s (%s)", strings.Title(dashboardName), setup.Peer, setup.OrgName))

		// Persist the created dashboard in the configured directory, from where it is going to be loaded
		err = ioutil.WriteFile(fmt.Sprintf("%s/%s-%s-%s.json", setup.DashboardDirectory, dashboardName, setup.Peer, setup.OrgName), []byte(dashboard), 0664)
		if err != nil {
			return err
		}
	}

	return nil
}
