package integration

import (
	"github.com/srabraham/ranger-ims-go/api"
	imsjson "github.com/srabraham/ranger-ims-go/json"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"
)

func sampleIncident1(eventName string) imsjson.Incident {
	return imsjson.Incident{
		Event:    eventName,
		State:    "new",
		Priority: 5,
		Summary:  ptr("my summary!"),
		Location: imsjson.Location{
			Name:         ptr("Zeroth Camp"),
			RadialHour:   ptr("10"),
			RadialMinute: ptr("5"),
			Description:  ptr("unknown"),
			Type:         "garett",
		},
		IncidentTypes: &[]string{"Admin", "Junk"},
		FieldReports:  &[]int32{},
		RangerHandles: &[]string{"SomeOne", "SomeTwo"},
		ReportEntries: []imsjson.ReportEntry{
			{Text: "This is some report text lol"},
		},
	}
}

func TestIncidentAPIAuthorization(t *testing.T) {
	s := httptest.NewServer(api.AddToMux(nil, shared.cfg, shared.imsDB, nil))
	defer s.Close()
	serverURL, err := url.Parse(s.URL)
	require.NoError(t, err)

	apisAdmin := ApiHelper{t: t, serverURL: serverURL, jwt: jwtForTestAdminRanger(t)}
	apisNonAdmin := ApiHelper{t: t, serverURL: serverURL, jwt: jwtForRealTestUser(t)}
	apisNotAuthenticated := ApiHelper{t: t, serverURL: serverURL, jwt: ""}

	// make an event, to which no one has access
	eventName := "IncidentEvent-943034"
	resp := apisAdmin.editEvent(imsjson.EditEventsRequest{Add: []string{eventName}})
	require.Equal(t, http.StatusNoContent, resp.StatusCode)

	// test all the APIs... getIncidents, getIncident, newIncident, updateIncident
	_, resp = apisNotAuthenticated.getIncidents(eventName)
	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	_, resp = apisNonAdmin.getIncidents(eventName)
	require.Equal(t, http.StatusForbidden, resp.StatusCode)
	_, resp = apisAdmin.getIncidents(eventName)
	require.Equal(t, http.StatusForbidden, resp.StatusCode)
	// get incident
	_, resp = apisNotAuthenticated.getIncident(eventName, 1)
	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	_, resp = apisNonAdmin.getIncident(eventName, 1)
	require.Equal(t, http.StatusForbidden, resp.StatusCode)
	_, resp = apisAdmin.getIncident(eventName, 1)
	require.Equal(t, http.StatusForbidden, resp.StatusCode)
	// new incident
	resp = apisNotAuthenticated.newIncident(imsjson.Incident{Event: eventName})
	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	resp = apisNonAdmin.newIncident(imsjson.Incident{Event: eventName})
	require.Equal(t, http.StatusForbidden, resp.StatusCode)
	resp = apisAdmin.newIncident(imsjson.Incident{Event: eventName})
	require.Equal(t, http.StatusForbidden, resp.StatusCode)
	// update incident
	resp = apisNotAuthenticated.updateIncident(eventName, 1, imsjson.Incident{})
	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	resp = apisNonAdmin.updateIncident(eventName, 1, imsjson.Incident{})
	require.Equal(t, http.StatusForbidden, resp.StatusCode)
	resp = apisAdmin.updateIncident(eventName, 1, imsjson.Incident{})
	require.Equal(t, http.StatusForbidden, resp.StatusCode)

	// make Alice a writer
	resp = apisAdmin.addWriter(eventName, userAliceHandle)
	require.Equal(t, http.StatusNoContent, resp.StatusCode)

	// test all the APIs again... getIncidents, getIncident, newIncident, updateIncident
	_, resp = apisNotAuthenticated.getIncidents(eventName)
	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	_, resp = apisNonAdmin.getIncidents(eventName)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	_, resp = apisAdmin.getIncidents(eventName)
	require.Equal(t, http.StatusForbidden, resp.StatusCode)
	// get incident
	_, resp = apisNotAuthenticated.getIncident(eventName, 1)
	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	_, resp = apisNonAdmin.getIncident(eventName, 1)
	require.Equal(t, http.StatusNotFound, resp.StatusCode)
	_, resp = apisAdmin.getIncident(eventName, 1)
	require.Equal(t, http.StatusForbidden, resp.StatusCode)
	// new incident
	resp = apisNotAuthenticated.newIncident(imsjson.Incident{Event: eventName})
	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	resp = apisNonAdmin.newIncident(imsjson.Incident{Event: eventName})
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	resp = apisAdmin.newIncident(imsjson.Incident{Event: eventName})
	require.Equal(t, http.StatusForbidden, resp.StatusCode)
	// update incident
	resp = apisNotAuthenticated.updateIncident(eventName, 1, imsjson.Incident{})
	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	resp = apisNonAdmin.updateIncident(eventName, 1, imsjson.Incident{})
	require.Equal(t, http.StatusNoContent, resp.StatusCode)
	resp = apisAdmin.updateIncident(eventName, 1, imsjson.Incident{})
	require.Equal(t, http.StatusForbidden, resp.StatusCode)
}

func TestCreateAndGetIncident(t *testing.T) {
	s := httptest.NewServer(api.AddToMux(nil, shared.cfg, shared.imsDB, nil))
	defer s.Close()
	serverURL, err := url.Parse(s.URL)
	require.NoError(t, err)

	apisAdmin := ApiHelper{t: t, serverURL: serverURL, jwt: jwtForTestAdminRanger(t)}
	apisNonAdmin := ApiHelper{t: t, serverURL: serverURL, jwt: jwtForRealTestUser(t)}

	// Use the admin JWT to create a new event,
	// then give the normal user Writer role on that event
	eventName := "IncidentEvent-1"
	resp := apisAdmin.editEvent(imsjson.EditEventsRequest{Add: []string{eventName}})
	require.Equal(t, http.StatusNoContent, resp.StatusCode)
	apisAdmin.addWriter(eventName, userAliceHandle)
	require.Equal(t, http.StatusNoContent, resp.StatusCode)

	// Use normal user to create a new Incident
	incidentReq := sampleIncident1(eventName)
	entryReq := incidentReq.ReportEntries[0]
	num := apisNonAdmin.newIncidentSuccess(incidentReq)
	incidentReq.Number = num

	// Use normal user to fetch that Incident from the API and check it over
	retrievedIncident, resp := apisNonAdmin.getIncident(eventName, num)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.NotNil(t, retrievedIncident)
	require.WithinDuration(t, time.Now(), retrievedIncident.Created, 5*time.Minute)
	require.WithinDuration(t, time.Now(), retrievedIncident.LastModified, 5*time.Minute)
	require.Len(t, retrievedIncident.ReportEntries, 2)

	// The first report entry will be the system entry. The second should be the one we sent in the request
	retrievedUserEntry := retrievedIncident.ReportEntries[1]
	retrievedUserEntry.ID = 0
	require.WithinDuration(t, time.Now(), retrievedUserEntry.Created, 5*time.Minute)
	retrievedUserEntry.Created = time.Time{}
	entryReq.Author = userAliceHandle
	require.Equal(t, entryReq, retrievedUserEntry)
	requireEqualIncident(t, incidentReq, retrievedIncident)

	// Now get the incident via the GetIncidents (plural) endpoint, and repeat the validation
	retrievedIncidents, resp := apisNonAdmin.getIncidents(eventName)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.Len(t, retrievedIncidents, 1)

	// The first entry will be the system entry. The second should be the one we sent in the request
	retrievedUserEntry = retrievedIncident.ReportEntries[1]
	retrievedUserEntry.ID = 0
	require.WithinDuration(t, time.Now(), retrievedUserEntry.Created, 5*time.Minute)
	retrievedUserEntry.Created = time.Time{}
	entryReq.Author = userAliceHandle
	require.Equal(t, entryReq, retrievedUserEntry)

	requireEqualIncident(t, incidentReq, retrievedIncidents[0])
}

func TestCreateAndUpdateIncident(t *testing.T) {
	s := httptest.NewServer(api.AddToMux(nil, shared.cfg, shared.imsDB, nil))
	defer s.Close()
	serverURL, err := url.Parse(s.URL)
	require.NoError(t, err)

	apisAdmin := ApiHelper{t: t, serverURL: serverURL, jwt: jwtForTestAdminRanger(t)}
	apisNonAdmin := ApiHelper{t: t, serverURL: serverURL, jwt: jwtForRealTestUser(t)}

	// Use the admin JWT to create a new event,
	// then give the normal user Writer role on that event
	eventName := "IncidentEvent-3829"
	resp := apisAdmin.editEvent(imsjson.EditEventsRequest{Add: []string{eventName}})
	require.Equal(t, http.StatusNoContent, resp.StatusCode)
	apisAdmin.addWriter(eventName, userAliceHandle)
	require.Equal(t, http.StatusNoContent, resp.StatusCode)

	// Use normal user to create a new Incident
	incidentReq := sampleIncident1(eventName)
	num := apisNonAdmin.newIncidentSuccess(incidentReq)
	incidentReq.Number = num

	retrievedNewIncident, resp := apisNonAdmin.getIncident(eventName, num)

	// Now let's update the incident. First let's try changing nothing.
	updates := imsjson.Incident{
		Event:  incidentReq.Event,
		Number: num,
	}

	resp = apisNonAdmin.updateIncident(eventName, num, updates)
	require.Equal(t, http.StatusNoContent, resp.StatusCode)

	retrievedIncidentAfterUpdate, resp := apisNonAdmin.getIncident(eventName, num)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	requireEqualIncident(t, retrievedNewIncident, retrievedIncidentAfterUpdate)

	// now let's set all fields to empty
	updates = imsjson.Incident{
		Event:    incidentReq.Event,
		Number:   num,
		State:    "closed",
		Priority: 1,
		Summary:  ptr(""),
		Location: imsjson.Location{
			Name:         ptr(""),
			Concentric:   ptr(""),
			RadialHour:   ptr(""),
			RadialMinute: ptr(""),
			Description:  ptr(""),
			Type:         "garett",
		},
		IncidentTypes: &[]string{},
		FieldReports:  &[]int32{},
		RangerHandles: &[]string{},
		ReportEntries: []imsjson.ReportEntry{},
	}
	resp = apisNonAdmin.updateIncident(eventName, num, updates)
	require.Equal(t, http.StatusNoContent, resp.StatusCode)
	// then check the result
	retrievedIncidentAfterUpdate, resp = apisNonAdmin.getIncident(eventName, num)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	expected := imsjson.Incident{
		Event:    eventName,
		Number:   num,
		State:    "closed",
		Priority: 1,
		Location: imsjson.Location{
			Type: "garett",
		},
		IncidentTypes: &[]string{},
		FieldReports:  &[]int32{},
		RangerHandles: &[]string{},
	}
	requireEqualIncident(t, expected, retrievedIncidentAfterUpdate)
}

// requireEqualIncident is a hacky way of checking two incident responses are the same.
// It does not consider ReportEntries.
func requireEqualIncident(t *testing.T, before imsjson.Incident, after imsjson.Incident) {
	// This field isn't in use in the client yet
	// require.Equal(t, before.EventID, after.EventID)
	require.Equal(t, before.Event, after.Event)
	require.Equal(t, before.Number, after.Number)

	// If the timestamp field was set before, then check it's the same. Otherwise
	// see if it was set to some reasonable time for when the test was running
	if !before.Created.IsZero() {
		require.Equal(t, before.Created, after.Created)
	} else {
		require.WithinDuration(t, time.Now(), after.Created, 20*time.Minute)
	}
	require.WithinDuration(t, time.Now(), after.LastModified, 20*time.Minute)
	require.Equal(t, before.State, after.State)
	require.Equal(t, before.Priority, after.Priority)
	require.Equal(t, before.Summary, after.Summary)
	require.Equal(t, before.Location, after.Location)
	require.Equal(t, before.IncidentTypes, after.IncidentTypes)
	require.Equal(t, before.RangerHandles, after.RangerHandles)
	require.Equal(t, before.FieldReports, after.FieldReports)
	// these will always be different. Check them separately of this function
	// require.Equal(t, before.ReportEntries, after.ReportEntries)

	//before.EventID = 0
	//after.EventID = 0
	//before.Created = time.Time{}
	//after.Created = time.Time{}
	//before.LastModified = time.Time{}
	//after.LastModified = time.Time{}
	//after.ReportEntries = nil
	//before.ReportEntries = nil
	//
	//require.Equal(t, before, after)
}

func ptr[T any](s T) *T {
	return &s
}
