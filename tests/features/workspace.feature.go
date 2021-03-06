package features

import (
	"cloud.google.com/go/firestore"
	"context"
	"encoding/json"
	"fmt"
	"github.com/cucumber/godog"
	"github.com/cucumber/godog/gherkin"
	"github.com/google/uuid"
	"github.com/metabs/server/customer"
	customerInt "github.com/metabs/server/internal/customer"
	"github.com/metabs/server/internal/jwt"
	workspaceInt "github.com/metabs/server/internal/workspace"
	"github.com/metabs/server/tab/collection/workspace"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

type workspaceFeature struct {
	db    *firestore.Client
	sv    *jwt.SignerVerifier
	res   *http.Response
	ws    *workspace.Workspace
	wss   []*workspace.Workspace
	body  []byte
	token []byte
}

func WorkspaceAPIs(s *godog.Suite, sv *jwt.SignerVerifier, db *firestore.Client) {
	f := &workspaceFeature{
		db:  db,
		sv:  sv,
		res: &http.Response{},
	}
	s.BeforeScenario(func(_ interface{}) {
		f.res = &http.Response{}
		f.ws = &workspace.Workspace{}
		f.wss = []*workspace.Workspace{}
		f.body = nil
		f.token = nil
	})
	s.Step(`^given the response body$`, f.givenResponseBody)
	s.Step(`^an existing workspace:$`, f.anExistingWorkspace)
	s.Step(`^an authenticated customer:$`, f.anAuthenticatedCustomer)
	s.Step(`^given the response body as list$`, f.givenResponseBodyAsList)
	s.Step(`^an HTTP "([^"]*)" request "([^"]*)":$`, f.anHTTPRequestWithTheURIAndBody)
	s.Step(`^the API must reply with a status code (\d+)$`, f.theAPIMustReplyWithAStatusCode)
	s.Step(`^the API must reply with a body containing:$`, f.theAPIMustReplyWithABodyContaining)
	s.Step(`^the API must reply with a body containing an id$`, f.theAPIMustReplyWithABodyContainingAnId)
	s.Step(`^the API must reply with a body containing an id as "([^"]*)"$`, f.theAPIMustReplyWithABodyContainingAnIdAs)
	s.Step(`^the API must reply with a body containing a name as "([^"]*)"$`, f.theAPIMustReplyWithABodyContainingANameAs)
	s.Step(`^the API must reply with a body containing nil update date$`, f.theAPIMustReplyWithABodyContainingNilUpdateDate)
	s.Step(`^the API must reply with a body containing an creation date$`, f.theAPIMustReplyWithABodyContainingAnCreationDate)
	s.Step(`^the API must reply with a body containing a customer id as "([^"]*)"$`, f.theAPIMustReplyWithABodyContainingACustomerIdAs)
	s.Step(`^the API must reply with a body containing an empty list of collections$`, f.theAPIMustReplyWithABodyContainingAnEmptyListOfCollections)
	s.Step(`^the API must reply with a body containing an update after create date$`, f.theAPIMustReplyWithABodyContainingAnUpdateDateAfterCreateDate)
	s.Step(`^the API must reply with a body containing a collections at index (\d+) containing an id$`, f.theAPIMustReplyWithABodyContainingACollectionsAtIndexContainingAnId)
	s.Step(`^the API must reply with a body containing a collections at index (\d+) containing an id as "([^"]*)"$`, f.theAPIMustReplyWithABodyContainingACollectionsAtIndexContainingAnIdAs)
	s.Step(`^the API must reply with a body containing a collections at index (\d+) containing a name as "([^"]*)"$`, f.theAPIMustReplyWithABodyContainingACollectionsAtIndexContainingANameAs)
	s.Step(`^the API must reply with a body containing a collections at index (\d+) containing nil update date$`, f.theAPIMustReplyWithABodyContainingACollectionsAtIndexContainingNilUpdateDate)
	s.Step(`^the API must reply with a body containing a collections at index (\d+) containing an creation date$`, f.theAPIMustReplyWithABodyContainingACollectionsAtIndexContainingAnCreationDate)
	s.Step(`^the API must reply with a body containing a collections at index (\d+) containing an update after create date$`, f.theAPIMustReplyWithABodyContainingACollectionsAtIndexContainingAnUpdateDateAfterCreateDate)
	s.Step(`^the API must reply with a body containing a collections at index (\d+) containing a tab at index (\d+) containing an id$`, f.theAPIMustReplyWithABodyContainingACollectionsAtIndexContainingATabAtIndexContainingAnId)
	s.Step(`^the API must reply with a body containing a collections at index (\d+) containing a tab at index (\d+) containing an id as "([^"]*)"$`, f.theAPIMustReplyWithABodyContainingACollectionsAtIndexContainingATabAtIndexContainingAnIdAs)
	s.Step(`^the API must reply with a body containing a collections at index (\d+) containing a tab at index (\d+) containing a icon as "([^"]*)"$`, f.theAPIMustReplyWithABodyContainingACollectionsAtIndexContainingATabAtIndexContainingAIconAs)
	s.Step(`^the API must reply with a body containing a collections at index (\d+) containing a tab at index (\d+) containing a link as "([^"]*)"$`, f.theAPIMustReplyWithABodyContainingACollectionsAtIndexContainingATabAtIndexContainingALinkAs)
	s.Step(`^the API must reply with a body containing a collections at index (\d+) containing a tab at index (\d+) containing a creation date$`, f.theAPIMustReplyWithABodyContainingACollectionsAtIndexContainingATabAtIndexContainingACreationDate)
	s.Step(`^the API must reply with a body containing a collections at index (\d+) containing a tab at index (\d+) containing nil update date$`, f.theAPIMustReplyWithABodyContainingACollectionsAtIndexContainingATabAtIndexContainingNilUpdateDate)
	s.Step(`^the API must reply with a body containing a collections at index (\d+) containing a tab at index (\d+) containing a title as "([^"]*)"$`, f.theAPIMustReplyWithABodyContainingACollectionsAtIndexContainingATabAtIndexContainingATitleAs)
	s.Step(`^the API must reply with a body containing a collections at index (\d+) containing a tab at index (\d+) containing a description as "([^"]*)"$`, f.theAPIMustReplyWithABodyContainingACollectionsAtIndexContainingATabAtIndexContainingADescriptionAs)
	s.Step(`^the API must reply with a body containing a collections at index (\d+) containing a tab at index (\d+) containing an update after create date$`, f.theAPIMustReplyWithABodyContainingACollectionsAtIndexContainingATabAtIndexContainingAnUpdateAfterCreateDate)

}

func (f *workspaceFeature) anHTTPRequestWithTheURIAndBody(method, uri string, body *gherkin.DocString) error {
	req, err := http.NewRequest(method, uri, strings.NewReader(body.Content))
	if err != nil {
		return fmt.Errorf("request creation failed. Due to: %s", err)
	}

	if string(f.token) != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", f.token))
	}
	client := http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("client failed. Due to: %s", err)
	}

	f.res = res
	return nil
}

func (f *workspaceFeature) theAPIMustReplyWithAStatusCode(expectedStatusCode int) error {

	if f.res.StatusCode != expectedStatusCode {
		return fmt.Errorf("response status cose is wrong. Expected: %d, Given: %d", http.StatusOK, f.res.StatusCode)
	}

	return nil
}

func (f *workspaceFeature) givenResponseBody() error {
	data, err := ioutil.ReadAll(f.res.Body)
	if err != nil {
		return fmt.Errorf("response body can't be read. Due to: %s", err)
	}
	var ws *workspace.Workspace
	if err := json.Unmarshal(data, &ws); err != nil {
		return fmt.Errorf("response body can't be unmarshalled. Due to: %s", err)
	}

	f.body = data
	f.ws = ws
	return nil
}

func (f *workspaceFeature) givenResponseBodyAsList() error {
	data, err := ioutil.ReadAll(f.res.Body)
	if err != nil {
		return fmt.Errorf("response body can't be read. Due to: %s", err)
	}
	var wss []*workspace.Workspace
	if err := json.Unmarshal(data, &wss); err != nil {
		return fmt.Errorf("response body can't be unmarshalled. Due to: %s", err)
	}

	f.body = data
	f.wss = wss
	return nil
}

func (f *workspaceFeature) theAPIMustReplyWithABodyContainingAnId() error {
	if _, err := uuid.Parse(string(f.ws.ID)); err != nil {
		return fmt.Errorf("id is wrong. Expected an uuid back, Given: %s", f.ws.ID)
	}

	return nil
}

func (f *workspaceFeature) theAPIMustReplyWithABodyContainingACustomerIdAs(id string) error {
	if f.ws.CustomerID.String() != id {
		return fmt.Errorf("id is wrong. Expected %s, Given: %s", id, f.ws.CustomerID)
	}

	return nil
}

func (f *workspaceFeature) theAPIMustReplyWithABodyContainingAnIdAs(id string) error {
	if string(f.ws.ID) != id {
		return fmt.Errorf("id is wrong. Expected %s, Given: %s", id, f.ws.ID)
	}

	return nil
}

func (f *workspaceFeature) theAPIMustReplyWithABodyContainingACollectionsAtIndexContainingAnId(i int) error {
	if _, err := uuid.Parse(string(f.ws.Collections[i].ID)); err != nil {
		return fmt.Errorf("id is wrong. Expected an uuid back, Given: %s", f.ws.ID)
	}

	return nil
}

func (f *workspaceFeature) theAPIMustReplyWithABodyContainingACollectionsAtIndexContainingAnIdAs(i int, id string) error {
	if string(f.ws.Collections[i].ID) != id {
		return fmt.Errorf("id is wrong. Expected %s, Given: %s", id, f.ws.ID)
	}

	return nil
}

func (f *workspaceFeature) theAPIMustReplyWithABodyContainingANameAs(wantName string) error {

	if string(f.ws.Name) != wantName {
		return fmt.Errorf("Name is wrong. Expected %s, Given: %s ", wantName, f.ws.Name)
	}

	return nil
}

func (f *workspaceFeature) theAPIMustReplyWithABodyContainingACollectionsAtIndexContainingANameAs(i int, wantName string) error {

	if string(f.ws.Collections[i].Name) != wantName {
		return fmt.Errorf("Name is wrong. Expected %s, Given: %s ", wantName, f.ws.Name)
	}

	return nil
}

func (f *workspaceFeature) theAPIMustReplyWithABodyContainingAnEmptyListOfCollections() error {
	if len(f.ws.Collections) > 0 {
		return fmt.Errorf("collection is wrong. Expected an empty collection back, Given: %v", f.ws.Collections)
	}

	return nil
}

func (f *workspaceFeature) theAPIMustReplyWithABodyContainingAnCreationDate() error {
	now := time.Now().Add(time.Second) // for safety
	if f.ws.Created.IsZero() || f.ws.Created.After(now) {
		return fmt.Errorf("creation date is wrong. Expected before than %v, Given: %v", now, f.ws.Created)
	}

	return nil
}

func (f *workspaceFeature) theAPIMustReplyWithABodyContainingNilUpdateDate() error {

	if !f.ws.Updated.IsZero() {
		return fmt.Errorf("update date is wrong. Expected a nil date, Given: %v", f.ws.Updated)
	}

	return nil
}

func (f *workspaceFeature) theAPIMustReplyWithABodyContainingAnUpdateDateAfterCreateDate() error {

	if !f.ws.Updated.After(f.ws.Created) {
		return fmt.Errorf("update date is wrong. Expected a after creation %v, Given: %v", f.ws.Created, f.ws.Updated)
	}

	return nil
}

func (f *workspaceFeature) theAPIMustReplyWithABodyContainingACollectionsAtIndexContainingAnCreationDate(i int) error {
	now := time.Now().Add(time.Second) // for safety
	if f.ws.Collections[i].Created.IsZero() || f.ws.Collections[i].Created.After(now) {
		return fmt.Errorf("creation date is wrong. Expected before than %v, Given: %v", now, f.ws.Collections[i].Created)
	}

	return nil
}

func (f *workspaceFeature) theAPIMustReplyWithABodyContainingACollectionsAtIndexContainingNilUpdateDate(i int) error {

	if !f.ws.Collections[i].Updated.IsZero() {
		return fmt.Errorf("update date is wrong. Expected a nil date, Given: %v", f.ws.Collections[i].Updated)
	}

	return nil
}

func (f *workspaceFeature) theAPIMustReplyWithABodyContainingACollectionsAtIndexContainingAnUpdateDateAfterCreateDate(i int) error {

	if !f.ws.Collections[i].Updated.After(f.ws.Collections[i].Created) {
		return fmt.Errorf("update date is wrong. Expected a after creation %v, Given: %v", f.ws.Collections[i].Created, f.ws.Collections[i].Updated)
	}

	return nil
}

func (f *workspaceFeature) anExistingWorkspace(data *gherkin.DocString) error {
	var ws workspace.Workspace
	if err := json.NewDecoder(strings.NewReader(data.Content)).Decode(&ws); err != nil {
		return fmt.Errorf("could not decode workspace. Due to: %s", err)
	}
	_, err := f.db.Collection(workspaceInt.CollectionName).Doc(string(ws.ID)).Set(context.Background(), ws)
	if err != nil {
		return fmt.Errorf("could not add workspace due to :%s", err)
	}

	return nil
}

func (f *workspaceFeature) anAuthenticatedCustomer(data *gherkin.DocString) error {
	var c customer.Customer
	if err := json.NewDecoder(strings.NewReader(data.Content)).Decode(&c); err != nil {
		return fmt.Errorf("could not decode workspace. Due to: %s", err)
	}
	_, err := f.db.Collection(customerInt.CollectionName).Doc(c.Email.String()).Set(context.Background(), c)
	if err != nil {
		return fmt.Errorf("could not add customer due to :%s", err)
	}
	token, err := f.sv.Sign(c.ID, c.Email, c.Status, c.Created, c.Activated, c.Updated)
	if err != nil {
		return fmt.Errorf("could not sign customer due to :%s", err)
	}

	f.token = token
	return nil
}

func (f *workspaceFeature) theAPIMustReplyWithABodyContaining(expectedData *gherkin.DocString) error {
	if strings.Trim(string(f.body), "\n") != strings.Trim(expectedData.Content, "\n") {
		return fmt.Errorf("workspace is wrong. Expected %s date, Given: %s", expectedData.Content, string(f.body))
	}

	return nil
}

func (f *workspaceFeature) theAPIMustReplyWithABodyContainingACollectionsAtIndexContainingATabAtIndexContainingAnId(collI, tabI int) error {
	if _, err := uuid.Parse(string(f.ws.Collections[collI].Tabs[tabI].ID)); err != nil {
		return fmt.Errorf("id is wrong. Expected an uuid back, Given: %s", f.ws.Collections[collI].Tabs[tabI].ID)
	}

	return nil
}

func (f *workspaceFeature) theAPIMustReplyWithABodyContainingACollectionsAtIndexContainingATabAtIndexContainingAnIdAs(collI, tabI int, wantID string) error {
	if string(f.ws.Collections[collI].Tabs[tabI].ID) != wantID {
		return fmt.Errorf("id is wrong. Expected %s, Given: %s", wantID, f.ws.Collections[collI].Tabs[tabI].ID)
	}

	return nil
}

func (f *workspaceFeature) theAPIMustReplyWithABodyContainingACollectionsAtIndexContainingATabAtIndexContainingATitleAs(collI, tabI int, wantTitle string) error {
	if string(f.ws.Collections[collI].Tabs[tabI].Title) != wantTitle {
		return fmt.Errorf("link is wrong. Expected %s, Given: %s", wantTitle, f.ws.Collections[collI].Tabs[tabI].Title)
	}

	return nil
}

func (f *workspaceFeature) theAPIMustReplyWithABodyContainingACollectionsAtIndexContainingATabAtIndexContainingADescriptionAs(collI, tabI int, wantDesc string) error {
	if string(f.ws.Collections[collI].Tabs[tabI].Description) != wantDesc {
		return fmt.Errorf("link is wrong. Expected %s, Given: %s", wantDesc, f.ws.Collections[collI].Tabs[tabI].Description)
	}

	return nil
}

func (f *workspaceFeature) theAPIMustReplyWithABodyContainingACollectionsAtIndexContainingATabAtIndexContainingAIconAs(collI, tabI int, wantIcon string) error {
	if string(f.ws.Collections[collI].Tabs[tabI].Icon) != wantIcon {
		return fmt.Errorf("link is wrong. Expected %s, Given: %s", wantIcon, f.ws.Collections[collI].Tabs[tabI].Icon)
	}

	return nil
}

func (f *workspaceFeature) theAPIMustReplyWithABodyContainingACollectionsAtIndexContainingATabAtIndexContainingALinkAs(collI, tabI int, wantLink string) error {
	if string(f.ws.Collections[collI].Tabs[tabI].Link) != wantLink {
		return fmt.Errorf("link is wrong. Expected %s, Given: %s", wantLink, f.ws.Collections[collI].Tabs[tabI].Link)
	}

	return nil
}

func (f *workspaceFeature) theAPIMustReplyWithABodyContainingACollectionsAtIndexContainingATabAtIndexContainingACreationDate(collI, tabI int) error {
	now := time.Now().Add(time.Second) // for safety
	if f.ws.Collections[collI].Tabs[tabI].Created.IsZero() || f.ws.Collections[collI].Tabs[tabI].Created.After(now) {
		return fmt.Errorf("creation date is wrong. Expected before than %v, Given: %v", now, f.ws.Collections[collI].Tabs[tabI].Created)
	}

	return nil
}

func (f *workspaceFeature) theAPIMustReplyWithABodyContainingACollectionsAtIndexContainingATabAtIndexContainingNilUpdateDate(collI, tabI int) error {
	if !f.ws.Collections[collI].Tabs[tabI].Updated.IsZero() {
		return fmt.Errorf("update date is wrong. Expected a nil date, Given: %v", f.ws.Collections[collI].Tabs[tabI].Updated)
	}

	return nil
}

func (f *workspaceFeature) theAPIMustReplyWithABodyContainingACollectionsAtIndexContainingATabAtIndexContainingAnUpdateAfterCreateDate(collI, tabI int) error {
	if !f.ws.Collections[collI].Tabs[tabI].Updated.After(f.ws.Collections[collI].Tabs[tabI].Created) {
		return fmt.Errorf("update date is wrong. Expected a after creation %v, Given: %v", f.ws.Collections[collI].Tabs[tabI].Created, f.ws.Collections[collI].Tabs[tabI].Updated)
	}

	return nil
}
