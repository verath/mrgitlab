package youtrack

import "testing"
import "encoding/json"

var issueJSON = `{
    "comment": [
        {
            "author": "Tester_Test",
            "authorFullName": "Tester Test",
            "created": 1503911905598,
            "deleted": false,
            "id": "72-6133",
            "issueId": "XYZ-996",
            "jiraId": null,
            "parentId": null,
            "permittedGroup": null,
            "replies": [],
            "shownForIssueAuthor": false,
            "text": "This is a problem!",
            "updated": null
        }
    ],
    "entityId": "71-9690",
    "field": [
        {
            "name": "projectShortName",
            "value": "XYZ"
        },
        {
            "name": "numberInProject",
            "value": "996"
        },
        {
            "name": "summary",
            "value": "Product X example code problems"
        },
        {
            "name": "description",
            "value": "==Description==\n- Solve issues. Product X is currently using Y for Z. This might not be an optimal solution as X is shipped to customers."
        },
        {
            "name": "created",
            "value": "1502968008698"
        },
        {
            "name": "updated",
            "value": "1505119826928"
        },
        {
            "name": "updaterName",
            "value": "petere"
        },
        {
            "name": "updaterFullName",
            "value": "Peter Eliasson"
        },
        {
            "name": "resolved",
            "value": "1505119826922"
        },
        {
            "name": "reporterName",
            "value": "jsmith"
        },
        {
            "name": "reporterFullName",
            "value": "John Smith"
        },
        {
            "name": "commentsCount",
            "value": "1"
        },
        {
            "name": "votes",
            "value": "0"
        },
        {
            "name": "links",
            "value": [
                {
                    "role": "subtask of",
                    "type": "Subtask",
                    "value": "XYZ-990"
                }
            ]
        },
        {
            "color": {
                "bg": "#cc6600",
                "fg": "white"
            },
            "name": "Type",
            "value": [
                "Task"
            ],
            "valueId": [
                "Task"
            ]
        },
        {
            "color": null,
            "name": "State",
            "value": [
                "Done"
            ],
            "valueId": [
                "Done"
            ]
        },
        {
            "color": null,
            "name": "Backlog Priority",
            "value": [
                "No backlog priority"
            ],
            "valueId": [
                "No backlog priority"
            ]
        },
        {
            "color": null,
            "name": "Sprint Priority",
            "value": [
                "No Priority"
            ],
            "valueId": [
                "No Priority"
            ]
        },
        {
            "color": null,
            "name": "Sprint",
            "value": [
                "S65"
            ],
            "valueId": [
                "S34"
            ]
        },
        {
            "name": "Assignee",
            "value": [
                {
                    "fullName": "Tester Test",
                    "value": "Tester_Test"
                }
            ]
        },
        {
            "color": null,
            "name": "Subsystem",
            "value": [
                "No subsystem"
            ],
            "valueId": [
                "No subsystem"
            ]
        },
        {
            "color": null,
            "name": "Affected Builds",
            "value": [
                "Unknown"
            ],
            "valueId": [
                "Unknown"
            ]
        },
        {
            "color": null,
            "name": "Affected versions",
            "value": [
                "Unknown"
            ],
            "valueId": [
                "Unknown"
            ]
        }
    ],
    "id": "XYZ-996",
    "jiraId": null,
    "tag": []
}`

func TestIssueUnmarshal(t *testing.T) {
	issue := &Issue{}
	if err := json.Unmarshal([]byte(issueJSON), issue); err != nil {
		t.Fatalf("unexpected error: %+v", err)
	}
}

func TestIssueFieldStringValue(t *testing.T) {
	issue := &Issue{}
	if err := json.Unmarshal([]byte(issueJSON), issue); err != nil {
		t.Fatalf("unexpected error: %+v", err)
	}
	actual, err := issue.FieldStringValue("updaterName")
	if err != nil {
		t.Fatalf("unexpected error: %+v", err)
	}
	expected := "petere"
	if actual != expected {
		t.Errorf("expected value to equal '%s', was: '%s'", expected, actual)
	}
	_, err = issue.FieldStringValue("NON-EXISTING")
	if err == nil {
		t.Error("expected an error when trying to access value of non-existing field")
	}
}
