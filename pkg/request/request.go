package request

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

func ReadJSON(req *http.Request, into interface{}) error {
	if err := json.NewDecoder(req.Body).Decode(into); err != nil {
		return errors.New(fmt.Sprintf("invalid request body: %s", err.Error()))
	}
	return nil
}
