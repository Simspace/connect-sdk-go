package connect

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/opentracing/opentracing-go"
	"github.com/stretchr/testify/assert"

	"github.com/1Password/connect-sdk-go/onepassword"
)

var validHost string
var validToken string
var defaultVault string

var mockHTTPClient *mockClient
var testClient *restClient

var requestCount int
var requestFail bool
var testUserAgent string

var testServerDefaultVersion = version{1, 3, 0}

const testItemUUID = "2a47aa139ef74d7ca17918035e"
const testVaultUUID = "5b52aa139ef74d7ca17918nmf8"

type mockClient struct {
	Dofunc func(req *http.Request) (*http.Response, error)
}

const testID = "4eh55wjehsta5f376ggwsplxs4"

func (mc *mockClient) Do(req *http.Request) (*http.Response, error) {
	resp, err := mc.Dofunc(req)
	if err != nil {
		return nil, err
	}
	if resp.Header.Get(VersionHeaderKey) == "" {
		resp.Header.Set(VersionHeaderKey, testServerDefaultVersion.String())
	}
	return resp, nil
}

func TestMain(m *testing.M) {
	validHost = "http://localhost:8080"
	validToken = "eyJhbGciOiJFUzI1NiIsInR5cCI6IkpXVCJ9.eyIxcGFzc3dvcmQuY29tL2F1dWlkIjoiR1RLVjVWUk5UUkVHUEVMWE41QlBSQTJHTjQiLCIxcGFzc3dvcmQuY29tL2Z0cyI6WyJ2YXVsdGFjY2VzcyJdLCIxcGFzc3dvcmQuY29tL3Rva2VuIjoidTFxMUNtWVhtbGR2YWZUa0lHTW8tLTJnazl5a180SkMiLCIxcGFzc3dvcmQuY29tL3Z0cyI6W3siYSI6MTU3MzA2NzIsInUiOiJvdGw2cjZudWdqNXdyNjNybmt3M3Y0cGJuYSJ9XSwiYXVkIjpbImNvbS4xcGFzc3dvcmQuc2VjcmV0c2VydmljZSJdLCJpYXQiOjE2MDMxMjg2NDIsImlzcyI6ImNvbS4xcGFzc3dvcmQuYjUiLCJqdGkiOiI2bjYyZHhyanBxZW00aGJ4d3dxdGJtNmpsZSIsInN1YiI6IkFWNFFORUM3UFJGREZFRTJJREpNM0NSSUNJIn0.-1shD95-qGYrHh3beH5nrfsV91BMp30Y9ipIwE6n4pw8Y9-2fR-gun27ShS9fHLJqW9xJZ-Eir1UEkiha2ucvA"
	defaultVault = "otl6r6nugj5wr63rnkw3v4pbna"
	testUserAgent = fmt.Sprintf(defaultUserAgent, SDKVersion)

	os.Setenv("OP_VAULT", defaultVault)
	os.Setenv("OP_CONNECT_HOST", validHost)
	os.Setenv("OP_CONNECT_TOKEN", validToken)

	mockHTTPClient = &mockClient{}

	testClient = &restClient{
		URL:       validHost,
		Token:     validToken,
		userAgent: testUserAgent,
		tracer:    opentracing.GlobalTracer(),
		client:    mockHTTPClient,
	}

	requestCount = 0
	requestFail = false

	os.Exit(m.Run())
}

func TestNewClientFromEnvironmentWithoutHost(t *testing.T) {
	os.Unsetenv("OP_CONNECT_HOST")
	defer os.Setenv("OP_CONNECT_HOST", validHost)
	_, err := NewClientFromEnvironment()
	if err == nil {
		t.Log("Expected client to fail")
		t.FailNow()
	}
}

func TestNewClientFromEnvironmentWithoutToken(t *testing.T) {
	os.Unsetenv("OP_CONNECT_TOKEN")
	defer os.Setenv("OP_CONNECT_TOKEN", validToken)
	_, err := NewClientFromEnvironment()
	if err == nil {
		t.Log("Expected client to fail")
		t.FailNow()
	}
}

func TestNewClientFromEnvironment(t *testing.T) {
	client, err := NewClientFromEnvironment()
	if err != nil {
		t.Logf("Unable to create client from environment: %q", err)
		t.FailNow()
	}

	restClient, ok := client.(*restClient)
	if !ok {
		t.Log("Unable to cast client to rest client. Was expecting restClient")
		t.FailNow()
	}

	if restClient.URL != validHost {
		t.Logf("Expected host of http://localhost:8080, got %q", restClient.URL)
		t.FailNow()
	}

	if restClient.Token != validToken {
		t.Logf("Expected valid token %q, got %q", validToken, restClient.Token)
		t.FailNow()
	}

	if restClient.userAgent != testUserAgent {
		t.Logf("Expected user-agent of %q, got %q", testUserAgent, restClient.userAgent)
		t.FailNow()
	}
}

func TestNewClient(t *testing.T) {
	client := NewClient(validHost, validToken)

	restClient, ok := client.(*restClient)
	if !ok {
		t.Log("Unable to cast client to rest client. Was expecting restClient")
		t.FailNow()
	}

	if restClient.URL != validHost {
		t.Logf("Expected host of http://localhost:8080, got %q", restClient.URL)
		t.FailNow()
	}

	if restClient.Token != validToken {
		t.Logf("Expected valid token %q, got %q", validToken, restClient.Token)
		t.FailNow()
	}

	if restClient.userAgent != testUserAgent {
		t.Logf("Expected user-agent of %q, got %q", testUserAgent, restClient.userAgent)
		t.FailNow()
	}
}

func TestNewClientWithUserAgent(t *testing.T) {
	client := NewClientWithUserAgent(validHost, validToken, "testSuite")

	restClient, ok := client.(*restClient)
	if !ok {
		t.Log("Unable to cast client to rest client. Was expecting restClient")
		t.FailNow()
	}

	if restClient.URL != validHost {
		t.Logf("Expected host of http://localhost:8080, got %q", restClient.URL)
		t.FailNow()
	}

	if restClient.Token != validToken {
		t.Logf("Expected valid token %q, got %q", validToken, restClient.Token)
		t.FailNow()
	}

	if restClient.userAgent != "testSuite" {
		t.Logf("Expected user-agent of %q, got %q", defaultUserAgent, restClient.userAgent)
		t.FailNow()
	}

}

func Test_restClient_GetVaults(t *testing.T) {
	mockHTTPClient.Dofunc = listVaults
	vaults, err := testClient.GetVaults()

	if err != nil {
		t.Logf("Unable to get vaults: %s", err.Error())
		t.FailNow()
	}

	if len(vaults) < 1 {
		t.Logf("Expected vaults to exist, found %d", len(vaults))
		t.FailNow()
	}
}

func Test_restClient_GetVault(t *testing.T) {
	expectedVault := &onepassword.Vault{
		Name:        "Test vault",
		Description: "Test Vault description",
		ID:          testID,
	}

	mockHTTPClient.Dofunc = getVault(expectedVault)
	vault, err := testClient.GetVault(expectedVault.ID)

	assert.Nil(t, err)
	assert.Equal(t, expectedVault, vault, "retrieved vault is not as expected")
}

func Test_restClient_GetVaultByID(t *testing.T) {
	expectedVault := &onepassword.Vault{
		Name:        "Test vault",
		Description: "Test Vault description",
		ID:          testID,
	}

	mockHTTPClient.Dofunc = getVault(expectedVault)
	vault, err := testClient.GetVaultByUUID(expectedVault.ID)

	assert.Nil(t, err)
	assert.Equal(t, expectedVault, vault, "retrieved vault is not as expected")
}

func Test_restClient_GetVaultByTitle(t *testing.T) {
	expectedVault := &onepassword.Vault{
		Name:        "Test vault",
		Description: "Test Vault description",
		ID:          testID,
	}

	mockHTTPClient.Dofunc = listVaults
	vault, err := testClient.GetVaultByTitle(expectedVault.Name)

	assert.Nil(t, err)
	assert.Equal(t, expectedVault, vault, "retrieved vault is not as expected")
}

func Test_restClient_GetVaultEmptyQuery(t *testing.T) {
	errResult := apiError(http.StatusNotFound, "Vault not found")
	mockHTTPClient.Dofunc = respondError(errResult)
	_, err := testClient.GetVault("")

	assert.EqualError(t, err, "Please provide either the vault name or its ID.")
}

func Test_restClient_GetVaultError(t *testing.T) {
	errResult := apiError(http.StatusNotFound, "Vault not found")
	mockHTTPClient.Dofunc = respondError(errResult)
	_, err := testClient.GetVault(testID)

	assert.ErrorIs(t, err, errResult)
}

func Test_restClient_GetVaultsByTitle(t *testing.T) {
	mockHTTPClient.Dofunc = listVaults
	vaults, err := testClient.GetVaultsByTitle("Test Vault")

	if err != nil {
		t.Logf("Unable to get vaults: %s", err.Error())
		t.FailNow()
	}

	if len(vaults) < 1 {
		t.Logf("Expected vaults to exist, found %d", len(vaults))
		t.FailNow()
	}
}

func Test_restClient_GetItem(t *testing.T) {
	mockHTTPClient.Dofunc = getItem
	item, err := testClient.GetItem(testID, testID)

	if err != nil {
		t.Logf("Unable to get items: %s", err.Error())
		t.FailNow()
	}

	if item == nil {
		t.Log("Expected 1 item to exist")
		t.FailNow()
	}
}

func Test_restClient_GetItemByUUID(t *testing.T) {
	mockHTTPClient.Dofunc = getItem
	item, err := testClient.GetItemByUUID(testID, testID)

	if err != nil {
		t.Logf("Unable to get items: %s", err.Error())
		t.FailNow()
	}

	if item == nil {
		t.Log("Expected 1 item to exist")
		t.FailNow()
	}
}

func Test_restClient_GetItemNotFound(t *testing.T) {
	errResult := apiError(http.StatusNotFound, "item not found")
	mockHTTPClient.Dofunc = respondError(errResult)
	item, err := testClient.GetItem(testID, testID)

	assert.ErrorIs(t, err, errResult)
	if item != nil {
		t.Log("Expected no items returns")
		t.FailNow()
	}
}

func Test_restClient_GetItems(t *testing.T) {
	mockHTTPClient.Dofunc = listItems
	items, err := testClient.GetItems(testID)

	if err != nil {
		t.Logf("Unable to get item: %s", err.Error())
		t.FailNow()
	}

	if len(items) != 1 {
		t.Logf("Expected 1 item to exist in vault, found %d", len(items))
		t.FailNow()
	}
}

func Test_restClient_GetItemsByTitle(t *testing.T) {
	mockHTTPClient.Dofunc = listItemsOrGetItem
	items, err := testClient.GetItemsByTitle("test-item", testVaultUUID)

	if err != nil {
		t.Logf("Unable to get item: %s", err.Error())
		t.FailNow()
	}

	if len(items) != 1 {
		t.Logf("Expected 1 item to exist in vault, found %d", len(items))
		t.FailNow()
	}

	assert.Equal(t, items[0].Title, "test-item")
	assert.NotEqual(t, len(items[0].Fields), 0)
	assert.Equal(t, items[0].Fields[0].Value, "wendy")
	assert.Equal(t, items[0].Fields[1].Value, "appleseed")
}

func Test_restClient_GetItemByTitle(t *testing.T) {
	defer reset()

	mockHTTPClient.Dofunc = getItemByID
	item, err := testClient.GetItemByTitle("test", testID)

	if err != nil {
		t.Logf("Unable to get item: %s", err.Error())
		t.FailNow()
	}

	if item == nil {
		t.Log("Expected 1 item to exist")
		t.FailNow()
	}

	assert.Equal(t, item.Fields[0].Value, "wendy")
	assert.Equal(t, item.Fields[1].Value, "appleseed")

}

func Test_restClient_GetItemByNonUniqueTitle(t *testing.T) {
	requestFail = true
	defer reset()

	mockHTTPClient.Dofunc = getItemByID
	item, err := testClient.GetItemByTitle("test", testID)

	if err == nil {
		t.Log("Expected too many items")
		t.FailNow()
	}

	if item != nil {
		t.Log("Expected no items returns")
		t.FailNow()
	}
}

func Test_restClient_CreateItem(t *testing.T) {
	mockHTTPClient.Dofunc = createItem
	item, err := testClient.CreateItem(generateItem(defaultVault), defaultVault)

	if err != nil {
		t.Logf("Unable to create items: %s", err.Error())
		t.FailNow()
	}

	if item == nil {
		t.Log("Expected 1 item to be created")
		t.FailNow()
	}
}

func Test_restClient_CreateItemError(t *testing.T) {
	errResult := apiError(http.StatusBadRequest, "Vault UUID required")
	mockHTTPClient.Dofunc = respondError(errResult)
	item, err := testClient.CreateItem(generateItem(defaultVault), defaultVault)

	assert.ErrorIs(t, err, errResult)
	if item != nil {
		t.Log("Expected item to not be created")
		t.FailNow()
	}
}

func Test_restClient_UpdateItem(t *testing.T) {
	mockHTTPClient.Dofunc = updateItem
	item, err := testClient.UpdateItem(generateItem(defaultVault), "")

	if err != nil {
		t.Logf("Unable to update item: %s", err.Error())
		t.FailNow()
	}

	if item == nil {
		t.Log("Expected 1 item to be updated")
		t.FailNow()
	}
}

func Test_restClient_UpdateItemError(t *testing.T) {
	errResult := apiError(http.StatusBadRequest, "Missing required field")
	mockHTTPClient.Dofunc = respondError(errResult)

	item, err := testClient.UpdateItem(generateItem(defaultVault), "")

	assert.ErrorIs(t, err, errResult)
	if item != nil {
		t.Log("Expected item to not be update")
		t.FailNow()
	}
}

func Test_restClient_DeleteItem(t *testing.T) {
	mockHTTPClient.Dofunc = deleteItem
	err := testClient.DeleteItem(generateItem(defaultVault), "")

	if err != nil {
		t.Logf("Unable to delete item: %s", err.Error())
		t.FailNow()
	}
}

func Test_restClient_DeleteItemById(t *testing.T) {
	mockHTTPClient.Dofunc = deleteItem
	err := testClient.DeleteItemByID(generateItem(defaultVault).ID, defaultVault)

	if err != nil {
		t.Logf("Unable to delete item: %s", err.Error())
		t.FailNow()
	}
}

func Test_restClient_DeleteItemByTitle(t *testing.T) {
	mockHTTPClient.Dofunc = deleteItemByTitle
	err := testClient.DeleteItemByTitle(generateItem(defaultVault).Title, defaultVault)

	if err != nil {
		t.Logf("Unable to delete item: %s", err.Error())
		t.FailNow()
	}
}

func Test_restClient_DeleteItemError(t *testing.T) {
	errResult := apiError(http.StatusNotFound, "Vault not found")
	mockHTTPClient.Dofunc = respondError(errResult)

	err := testClient.DeleteItem(generateItem(defaultVault), "")

	assert.ErrorIs(t, err, errResult)
}

func Test_restClient_DeleteItemByIdError(t *testing.T) {
	errResult := apiError(http.StatusNotFound, "Vault not found")
	mockHTTPClient.Dofunc = respondError(errResult)

	err := testClient.DeleteItemByID(generateItem(defaultVault).ID, defaultVault)

	assert.ErrorIs(t, err, errResult)
}

func Test_restClient_GetFile(t *testing.T) {
	mockHTTPClient.Dofunc = getFile
	file, err := testClient.GetFile(testID, testID, testID)

	assert.Nil(t, err)
	assert.NotNil(t, file)
}

func Test_restClient_GetFileNotFound(t *testing.T) {
	errResult := apiError(http.StatusNotFound, "File not found")
	mockHTTPClient.Dofunc = respondError(errResult)
	_, err := testClient.GetFile(testID, testID, testID)

	assert.ErrorIs(t, err, errResult)
}

func Test_restClient_GetFileContent(t *testing.T) {
	f := generateFile()

	mockHTTPClient.Dofunc = getFileContent
	content, err := testClient.GetFileContent(f)

	assert.Nil(t, err)
	assert.Equal(t, []byte("test"), content)
}

func Test_restClient_GetFileContentError(t *testing.T) {
	f := generateFile()

	errResult := apiError(http.StatusNotFound, "File not found")
	mockHTTPClient.Dofunc = respondError(errResult)
	_, err := testClient.GetFileContent(f)

	assert.ErrorIs(t, err, errResult)
}

func Test_restClient_loadStructFromItem(t *testing.T) {
	type testConfig struct {
		Username string                  `opfield:"username"`
		Password string                  `opsection:"section" opfield:"password"`
		Section  onepassword.ItemSection `opsection:"section"`
		URL      onepassword.ItemURL     `opurl:"url"`
	}
	mockHTTPClient.Dofunc = getComplexItem

	item := parsedItem{
		vaultUUID: testID,
		itemUUID:  testID,
	}
	c := testConfig{}

	err := loadToStruct(&item, reflect.ValueOf(&c).Elem())
	assert.Nil(t, err)
	err = setValuesForTag(testClient, &item, false)
	assert.Nil(t, err)

	assert.Equal(t, "wendy", c.Username)
	assert.Equal(t, "appleseed", c.Password)
	assert.Equal(t, onepassword.ItemSection{
		ID:    "",
		Label: "section",
	}, c.Section)
	assert.Equal(t, onepassword.ItemURL{
		Label: "url",
		URL:   "https://www.appleseed.com",
	}, c.URL)
}

func Test_restClient_loadStructFromItemAndOneItem(t *testing.T) {
	type testConfig struct {
		Item     onepassword.Item        `opvault:"5b52aa139ef74d7ca17918nmf8" opitem:"test-item"`
		Username string                  `opfield:"username"`
		Password string                  `opsection:"section" opfield:"password"`
		Section  onepassword.ItemSection `opsection:"section"`
		URL      onepassword.ItemURL     `opurl:"url"`
	}
	mockHTTPClient.Dofunc = getComplexItem

	item := parsedItem{
		vaultUUID: testID,
		itemUUID:  testID,
	}
	c := testConfig{}

	err := loadToStruct(&item, reflect.ValueOf(&c).Elem())
	assert.Nil(t, err)
	err = setValuesForTag(testClient, &item, false)
	assert.Nil(t, err)

	assert.Equal(t, "wendy", c.Username)
	assert.Equal(t, "appleseed", c.Password)
	assert.Equal(t, onepassword.ItemSection{
		ID:    "",
		Label: "section",
	}, c.Section)
	assert.Equal(t, onepassword.ItemURL{
		Label: "url",
		URL:   "https://www.appleseed.com",
	}, c.URL)
	assert.Equal(t, generateComplexItem(testVaultUUID), c.Item)
}

func respondError(apiErr *onepassword.Error) func(req *http.Request) (*http.Response, error) {
	return func(req *http.Request) (*http.Response, error) {
		body, err := json.Marshal(apiErr)
		if err != nil {
			panic(err)
		}
		return &http.Response{
			Status:     http.StatusText(apiErr.StatusCode),
			StatusCode: apiErr.StatusCode,
			Header:     req.Header,
			Body:       ioutil.NopCloser(bytes.NewReader(body)),
		}, nil
	}
}

func listVaults(req *http.Request) (*http.Response, error) {
	vaults := []onepassword.Vault{
		{
			Name:        "Test vault",
			Description: "Test Vault description",
			ID:          testID,
		},
	}

	json, _ := json.Marshal(vaults)
	return &http.Response{
		Status:     http.StatusText(http.StatusOK),
		StatusCode: http.StatusOK,
		Body:       ioutil.NopCloser(bytes.NewReader(json)),
		Header:     req.Header,
	}, nil
}

func getVault(vault *onepassword.Vault) func(req *http.Request) (*http.Response, error) {
	return func(req *http.Request) (*http.Response, error) {
		json, _ := json.Marshal(vault)
		return &http.Response{
			Status:     http.StatusText(http.StatusOK),
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(bytes.NewReader(json)),
			Header:     req.Header,
		}, nil
	}
}

func generateComplexItem(vaultUUID string) onepassword.Item {
	return onepassword.Item{
		Title: "test-item",
		ID:    testItemUUID,
		Vault: onepassword.ItemVault{
			ID: vaultUUID,
		},
		Sections: []*onepassword.ItemSection{{
			ID:    "",
			Label: "section",
		}},
		Fields: []*onepassword.ItemField{
			{
				ID:    testID,
				Label: "username",
				Value: "wendy",
			},
			{
				Label: "password",
				Value: "appleseed",
				Section: &onepassword.ItemSection{
					ID:    "",
					Label: "section",
				},
			},
		},
		URLs: []onepassword.ItemURL{
			{
				Primary: true,
				Label:   "site",
				URL:     "https://www.wendy.com",
			},
			{
				Label: "url",
				URL:   "https://www.appleseed.com",
			},
		},
	}
}

func generateItem(vaultUUID string) *onepassword.Item {
	item := generateComplexItem(vaultUUID)
	return &item
}

func listItemsOrGetItem(req *http.Request) (*http.Response, error) {
	if strings.Contains(req.URL.RequestURI(), "test-item") {
		json, _ := json.Marshal([]onepassword.Item{generateComplexItem(testVaultUUID)})
		return &http.Response{
			Status:     http.StatusText(http.StatusOK),
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(bytes.NewReader(json)),
			Header:     req.Header,
		}, nil
	} else {
		json, _ := json.Marshal(generateComplexItem(testVaultUUID))
		return &http.Response{
			Status:     http.StatusText(http.StatusOK),
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(bytes.NewReader(json)),
			Header:     req.Header,
		}, nil
	}
}

func listItems(req *http.Request) (*http.Response, error) {
	vaultUUID := ""
	excessPath := ""
	fmt.Sscanf(req.URL.Path, "/v1/vaults/%s%s", vaultUUID, excessPath)

	items := []*onepassword.Item{
		generateItem(vaultUUID),
	}

	json, _ := json.Marshal(items)
	return &http.Response{
		Status:     http.StatusText(http.StatusOK),
		StatusCode: http.StatusOK,
		Body:       ioutil.NopCloser(bytes.NewReader(json)),
		Header:     req.Header,
	}, nil
}

func getItemByID(req *http.Request) (*http.Response, error) {
	vaultUUID := testID
	excessPath := ""
	fmt.Sscanf(req.URL.Path, "/v1/vaults/%s%s", vaultUUID, excessPath)

	items := []*onepassword.Item{
		generateItem(vaultUUID),
	}

	if requestFail {
		items = append(items, generateItem(vaultUUID))
	}

	if requestCount == 0 {
		requestCount++
		json, _ := json.Marshal(items)
		return &http.Response{
			Status:     http.StatusText(http.StatusOK),
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(bytes.NewReader(json)),
			Header:     req.Header,
		}, nil
	}

	return getItem(req)
}

func getComplexItem(req *http.Request) (*http.Response, error) {
	vaultUUID := testVaultUUID
	excessPath := ""
	fmt.Sscanf(req.URL.Path, "/v1/vaults/%s%s", vaultUUID, excessPath)

	json, _ := json.Marshal(generateComplexItem(vaultUUID))
	return &http.Response{
		Status:     http.StatusText(http.StatusOK),
		StatusCode: http.StatusOK,
		Body:       ioutil.NopCloser(bytes.NewReader(json)),
		Header:     req.Header,
	}, nil
}

func getItem(req *http.Request) (*http.Response, error) {
	vaultUUID := ""
	excessPath := ""
	fmt.Sscanf(req.URL.Path, "/v1/vaults/%s%s", vaultUUID, excessPath)

	json, _ := json.Marshal(generateItem(vaultUUID))
	return &http.Response{
		Status:     http.StatusText(http.StatusOK),
		StatusCode: http.StatusOK,
		Body:       ioutil.NopCloser(bytes.NewReader(json)),
		Header:     req.Header,
	}, nil
}

func createItem(req *http.Request) (*http.Response, error) {
	rawBody, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return nil, err
	}

	var item onepassword.Item
	if err := json.Unmarshal(rawBody, &item); err != nil {
		return nil, err
	}

	item.ID = testItemUUID
	item.CreatedAt = time.Now()

	vaultUUID := ""
	excessPath := ""
	fmt.Sscanf(req.URL.Path, "/v1/vaults/%s%s", vaultUUID, excessPath)

	item.Vault.ID = vaultUUID

	item.UpdatedAt = time.Now()

	json, _ := json.Marshal(item)
	return &http.Response{
		Status:     http.StatusText(http.StatusOK),
		StatusCode: http.StatusOK,
		Body:       ioutil.NopCloser(bytes.NewReader(json)),
		Header:     req.Header,
	}, nil
}

func updateItem(req *http.Request) (*http.Response, error) {
	rawBody, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return nil, err
	}

	var item onepassword.Item
	if err := json.Unmarshal(rawBody, &item); err != nil {
		return nil, err
	}

	item.UpdatedAt = time.Now()

	json, _ := json.Marshal(item)
	return &http.Response{
		Status:     http.StatusText(http.StatusOK),
		StatusCode: http.StatusOK,
		Body:       ioutil.NopCloser(bytes.NewReader(json)),
		Header:     req.Header,
	}, nil
}

func deleteItem(req *http.Request) (*http.Response, error) {
	vaultUUID := strings.ToLower(testVaultUUID)
	itemUUID := strings.ToLower(testItemUUID)
	fmt.Sscanf(req.URL.Path, "/v1/vaults/%s/items/%s", vaultUUID, itemUUID)

	return &http.Response{
		Status:     http.StatusText(http.StatusNoContent),
		StatusCode: http.StatusNoContent,
		Header:     req.Header,
		Body:       ioutil.NopCloser(&bytes.Buffer{}),
	}, nil
}

func deleteItemByTitle(req *http.Request) (*http.Response, error) {
	if req.Method != http.MethodDelete {
		return listItemsOrGetItem(req)
	}

	vaultUUID := strings.ToLower(testVaultUUID)
	itemUUID := strings.ToLower(testItemUUID)
	fmt.Sscanf(req.URL.Path, "/v1/vaults/%s/items/%s", vaultUUID, itemUUID)

	return &http.Response{
		Status:     http.StatusText(http.StatusNoContent),
		StatusCode: http.StatusNoContent,
		Header:     req.Header,
		Body:       ioutil.NopCloser(&bytes.Buffer{}),
	}, nil
}

func generateFile() *onepassword.File {
	return &onepassword.File{
		ID:          testID,
		Name:        "testfile.txt",
		ContentPath: "/v1/files/xbqdtnehinocwuz23c7l7jiagy/content",
	}
}

func getFile(req *http.Request) (*http.Response, error) {
	vaultUUID := ""
	itemUUID := ""
	excessPath := ""
	fmt.Sscanf(req.URL.Path, "/v1/vaults/%s/items/%s/files%s", vaultUUID, itemUUID, excessPath)

	json, _ := json.Marshal(generateFile())
	return &http.Response{
		Status:     http.StatusText(http.StatusOK),
		StatusCode: http.StatusOK,
		Body:       ioutil.NopCloser(bytes.NewReader(json)),
		Header:     req.Header,
	}, nil
}

func getFileContent(req *http.Request) (*http.Response, error) {
	fileUUID := ""
	excessPath := ""
	fmt.Sscanf(req.URL.Path, "/v1/files/%s%s", fileUUID, excessPath)

	return &http.Response{
		Status:     http.StatusText(http.StatusOK),
		StatusCode: http.StatusOK,
		Body:       ioutil.NopCloser(bytes.NewReader([]byte("test"))),
		Header:     req.Header,
	}, nil
}

func apiError(statusCode int, message string) *onepassword.Error {
	return &onepassword.Error{
		StatusCode: statusCode,
		Message:    message,
	}
}

func reset() {
	requestCount = 0
	requestFail = false
}
