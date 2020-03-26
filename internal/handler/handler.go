package handler

import (
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

var AllExpense = make(map[string]jobState)
var TokenMap = make(map[string][]byte)

//CallbackHandler ...
func CallbackHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	currState, ok := AllExpense[id]
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
	TokenMap[id] = []byte(taskToken)
	fmt.Fprint(w, "SUCCEED")
}

//ListHandler ...
func ListHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println(AllExpense)
	fmt.Fprint(w, "<h1>Job Processing SYSTEM</h1>"+"<a href=\"/workflow/job/list\">HOME</a>"+
		"<h3>All Job requests:</h3><table border=1><tr><th>Job ID</th><th>Status</th><th>Action</th>")
	keys := []string{}
	for k := range AllExpense {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, id := range keys {
		state := AllExpense[id]
		actionLink := ""
		if state == created {
			actionLink = fmt.Sprintf("<a href=\"/workflow/job/action?type=approve&id=%s\">"+
				"<button style=\"background-color:#4CAF50;\">APPROVE</button></a>"+
				"&nbsp;&nbsp;<a href=\"/workflow/job/action?type=reject&id=%s\">"+
				"<button style=\"background-color:#f44336;\">REJECT</button></a>", id, id)
		}
		fmt.Fprintf(w, "<tr><td>%s</td><td>%s</td><td>%s</td></tr>", id, state, actionLink)
	}
	fmt.Fprint(w, "</table>")
}

//StartJobHandler ...
func StartJobHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("inside StartJobHandler")
	isAPICall := r.URL.Query().Get("is_api_call") == "true"
	id := r.URL.Query().Get("id")
	_, ok := AllExpense[id]
	if ok {
		fmt.Fprint(w, "ERROR:ID_ALREADY_EXISTS")
		return
	}

	AllExpense[id] = created
	if isAPICall {
		fmt.Fprint(w, "SUCCEED")
	} else {
		ListHandler(w, r)
	}
	fmt.Printf("Created new job id:%s.\n", id)
	return
}

//ActionHandler ....
// func ActionHandler(w http.ResponseWriter, r *http.Request) {
// 	isAPICall := r.URL.Query().Get("is_api_call") == "true"
// 	id := r.URL.Query().Get("id")
// 	oldState, ok := AllExpense[id]
// 	if !ok {
// 		fmt.Fprint(w, "ERROR:INVALID_ID")
// 		return
// 	}
// 	actionType := r.URL.Query().Get("type")
// 	switch actionType {
// 	case "approve":
// 		AllExpense[id] = approved
// 	case "reject":
// 		AllExpense[id] = rejected
// 	case "processed":
// 		AllExpense[id] = completed
// 	}
// 	if isAPICall {
// 		fmt.Fprint(w, "SUCCEED")
// 	} else {
// 		ListHandler(w, r)
// 	}

// 	if oldState == created && (AllExpense[id] == approved || AllExpense[id] == rejected) {
// 		token, ok := TokenMap[id]
// 		if !ok {
// 			fmt.Printf("Invalid id:%s\n", id)
// 			return
// 		}
// 		var jobProcessorService *service.JobProcessorService
// 		err := jobProcessorService.CadenceAdapter.CadenceClient.CompleteActivity(context.Background(), token, state, nil)
// 		if err != nil {
// 			fmt.Printf("Failed to complete activity with error: %+v\n", err)
// 		} else {
// 			fmt.Printf("Successfully complete activity: %s\n", token)
// 		}
// 		// report state change
// 		// l.NotifyJobStateChange(id, string(allExpense[id]))
// 	}

// 	fmt.Printf("Set state for %s from %s to %s.\n", id, oldState, AllExpense[id])
// 	return
// }

//NotifyJobStateChange ...
// func (l *service.JobProcessorService) NotifyJobStateChange(id, state string) {
// 	token, ok := tokenMap[id]
// 	if !ok {
// 		fmt.Printf("Invalid id:%s\n", id)
// 		return
// 	}
// 	err := l.CadenceAdapter.CadenceClient.CompleteActivity(context.Background(), token, state, nil)
// 	if err != nil {
// 		fmt.Printf("Failed to complete activity with error: %+v\n", err)
// 	} else {
// 		fmt.Printf("Successfully complete activity: %s\n", token)
// 	}
// }
