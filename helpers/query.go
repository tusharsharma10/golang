package helpers

import (
	"errors"
	"fmt"
	"strings"

	"github.com/Knetic/govaluate"
)

func CreateInsertQuery(table string, columns []string) string {
	queryColumns := " (" + strings.Join(columns, ", ") + ") "
	queryValues := " ( :" + strings.Join(columns, ", :") + ") "

	return "INSERT INTO " + table + queryColumns + "VALUES" + queryValues
}

type QueryClause struct {
	InClause string
	Args     []interface{}
}

func PrepareInClauseQuery(data []string) (*QueryClause, error) {
	if len(data) == 0 {
		return nil, errors.New("empty data array")
	}

	inClause := make([]string, 0)
	qc := QueryClause{
		Args: make([]interface{}, 0),
	}

	for i := 0; i < len(data); i++ {
		inClause = append(inClause, "?")
		qc.Args = append(qc.Args, data[i])
	}

	qc.InClause = " ( " + strings.Join(inClause, ", ") + " ) "

	return &qc, nil
}

func PrepareIntegerInClauseQuery(data []int) (*QueryClause, error) {
	if len(data) == 0 {
		return nil, errors.New("empty data array")
	}

	inClause := make([]string, 0)
	qc := QueryClause{
		Args: make([]interface{}, 0),
	}

	for i := 0; i < len(data); i++ {
		inClause = append(inClause, "?")
		qc.Args = append(qc.Args, data[i])
	}

	qc.InClause = " ( " + strings.Join(inClause, ", ") + " ) "

	return &qc, nil
}

func ComputeExpression(expression string) (int, error) {
	evalExpression, err := govaluate.NewEvaluableExpression(expression)
	if err != nil {
		return 0, err
	}

	result, err := evalExpression.Evaluate(nil)
	if err != nil {
		return 0, err
	}

	resultCal, ok := result.(float64)
	if !ok {
		return 0, fmt.Errorf("unable to convert result to float")
	}

	return int(resultCal), nil
}

func Min(a, b int) int {
	if a < b {
		return a
	}

	return b
}
