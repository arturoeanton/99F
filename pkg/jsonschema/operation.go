package jsonschema

import (
	"strconv"
	"strings"

	"github.com/arturoeanton/99F/pkg/parser"
	"github.com/couchbase/gocb/v2"
	"github.com/google/uuid"
)

type ListResponse struct {
	TotalResults int           `json:"totalResults,omitempty"`
	ItemsPerPage int           `json:"itemsPerPage,omitempty"`
	StartIndex   int           `json:"startIndex,omitempty"`
	Resources    []interface{} `json:"resources,omitempty"`
}

func Create(data map[string]interface{}, nameSchema string) (map[string]interface{}, error) {
	id := uuid.New().String()
	return Replace(data, id, nameSchema)
}

func Replace(data map[string]interface{}, id string, nameSchema string) (map[string]interface{}, error) {
	element := make(map[string]interface{})
	element["meta"] = map[string]string{
		"id": id,
	}
	element["data"] = data
	bucket := Cluster.Bucket(nameSchema)
	collection := bucket.DefaultCollection()
	_, err := collection.Upsert(id, element, &gocb.UpsertOptions{})
	return element, err
}

func GetElementByID(id string, nameSchema string) (map[string]interface{}, error) {
	bucket := Cluster.Bucket(nameSchema)
	raw, err := bucket.DefaultCollection().Get(id, &gocb.GetOptions{})
	if err != nil {
		return nil, err
	}
	var result interface{}
	raw.Content(&result)
	data := result.(map[string]interface{})["data"].(map[string]interface{})
	return data, nil
}

func Remove(id, nameSchema string) error {
	bucket := Cluster.Bucket(nameSchema)
	collection := bucket.DefaultCollection()
	_, err := collection.Remove(id, &gocb.RemoveOptions{})
	if err != nil {
		return err
	}
	return nil
}

func Search(nameSchema, filter, startIndex, count, sortBy, sortOrder string) (*ListResponse, error) {
	var result ListResponse
	queryPage, queryCount := parser.FilterToN1QL(nameSchema, filter)

	if sortBy == "" {
		sortBy = "id"
	} else {
		sortByArray := strings.Split(sortBy, ",")
		cache := make([]string, 0)
		for _, s := range sortByArray {
			cache = append(cache, parser.AddQuote(s))
		}
		sortBy = strings.Join(cache, ",")
	}

	sortBy = strings.Trim(sortBy, " ")
	sortBy = strings.ReplaceAll(sortBy, ";", "")

	if sortOrder == "descending" {
		sortOrder = "DESC"
	} else {
		sortOrder = "ASC"
	}

	//pagination
	if startIndex == "" {
		startIndex = "1"
	}
	if count == "" {
		count = "100"
	}
	var err error
	result.StartIndex, err = strconv.Atoi(startIndex)
	if err != nil {
		return nil, err
	}
	if result.StartIndex < 1 {
		result.StartIndex = 1
	}
	result.ItemsPerPage, err = strconv.Atoi(count)
	if err != nil {
		return nil, err
	}
	queryPage += "\nORDER BY " + sortBy + " " + sortOrder
	queryPage += "\nOFFSET " + strconv.Itoa(result.StartIndex-1)
	queryPage += "\nLIMIT " + count

	//log.Println(queryCount)
	rowsCount, err := Cluster.Query(queryCount, nil)
	if err != nil {
		return nil, err
	}
	defer rowsCount.Close()

	var countResult struct {
		Count int
	}
	rowsCount.One(&countResult)
	//log.Println(queryPage)
	rows, err := Cluster.Query(queryPage, nil)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result.TotalResults = countResult.Count
	result.Resources = make([]interface{}, 0)
	for rows.Next() {
		var item map[string]interface{}
		err := rows.Row(&item)
		if err != nil {
			return nil, err
		}
		result.Resources = append(result.Resources, item[nameSchema])
	}
	return &result, nil
}
