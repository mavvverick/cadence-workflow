package handler

import (
	"context"
	"fmt"
	"net/http"
	"sort"
)

type jobState string

const (
	created   jobState = "CREATED"
	approved           = "APPROVED"
	rejected           = "REJECTED"
	completed          = "COMPLETED"
)

var allExpense = make(map[string]jobState)
var tokenMap = make(map[string][]byte)

//CallbackHandler ...
func CallbackHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	currState, ok := allExpense[id]
	if !ok {
		fmt.Fprint(w, "ERROR:INVALID_ID")
		return
	}
	if currState != created {
		fmt.Fprint(w, "ERROR:INVALID_STATE")
		return
	}

	err := r.ParseForm()
	if err != nil {
		// Handle error here via logging and then return
		fmt.Fprint(w, "ERROR:INVALID_FORM_DATA")
		return
	}

	taskToken := r.PostFormValue("task_token")
	fmt.Printf("Registered callback for ID=%s, token=%s\n", id, taskToken)
	tokenMap[id] = []byte(taskToken)
	fmt.Fprint(w, "SUCCEED")
}

//ListHandler ...
func ListHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "<h1>Job Processing SYSTEM</h1>"+"<a href=\"/cadence/job/list\">HOME</a>"+
		"<h3>All Job requests:</h3><table border=1><tr><th>Job ID</th><th>Status</th><th>Action</th>")
	keys := []string{}
	for k := range allExpense {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, id := range keys {
		state := allExpense[id]
		actionLink := ""
		if state == created {
			actionLink = fmt.Sprintf("<a href=\"/cadence/job/action?type=approve&id=%s\">"+
				"<button style=\"background-color:#4CAF50;\">APPROVE</button></a>"+
				"&nbsp;&nbsp;<a href=\"/cadence/job/action?type=reject&id=%s\">"+
				"<button style=\"background-color:#f44336;\">REJECT</button></a>", id, id)
		}
		fmt.Fprintf(w, "<tr><td>%s</td><td>%s</td><td>%s</td></tr>", id, state, actionLink)
	}
	fmt.Fprint(w, "</table>")
}

//CreateJobHandler ...
func CreateJobHandler(w http.ResponseWriter, r *http.Request) {
	isAPICall := r.URL.Query().Get("is_api_call") == "true"
	id := r.URL.Query().Get("id")
	_, ok := allExpense[id]
	if ok {
		fmt.Fprint(w, "ERROR:ID_ALREADY_EXISTS")
		return
	}

	allExpense[id] = created
	if isAPICall {
		fmt.Fprint(w, "SUCCEED")
	} else {
		ListHandler(w, r)
	}
	fmt.Printf("Created new job id:%s.\n", id)
	return
}

//ActionHandler ....
func (b *Service) ActionHandler(w http.ResponseWriter, r *http.Request) {
	isAPICall := r.URL.Query().Get("is_api_call") == "true"
	id := r.URL.Query().Get("id")
	oldState, ok := allExpense[id]
	if !ok {
		fmt.Fprint(w, "ERROR:INVALID_ID")
		return
	}
	actionType := r.URL.Query().Get("type")
	switch actionType {
	case "approve":
		allExpense[id] = approved
	case "reject":
		allExpense[id] = rejected
	case "processed":
		allExpense[id] = completed
	}
	if isAPICall {
		fmt.Fprint(w, "SUCCEED")
	} else {
		ListHandler(w, r)
	}

	if oldState == created && (allExpense[id] == approved || allExpense[id] == rejected) {
		// report state change
		b.NotifyJobStateChange(id, string(allExpense[id]))
	}

	fmt.Printf("Set state for %s from %s to %s.\n", id, oldState, allExpense[id])
	return
}

//NotifyJobStateChange ...
func (b *Service) NotifyJobStateChange(id, state string) {
	token, ok := tokenMap[id]
	if !ok {
		fmt.Printf("Invalid id:%s\n", id)
		return
	}
	err := b.CadenceAdapter.CadenceClient.CompleteActivity(context.Background(), token, state, nil)
	if err != nil {
		fmt.Printf("Failed to complete activity with error: %+v\n", err)
	} else {
		fmt.Printf("Successfully complete activity: %s\n", token)
	}
}
