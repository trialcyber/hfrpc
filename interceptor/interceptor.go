package interceptor

import "fmt"

type HandlerFunc func(*map[string]interface{}) error

func Token() HandlerFunc {
	return func(content *map[string]interface{}) error {
		if _, ok := (*content)["token"]; !ok {
			return fmt.Errorf("token is empty")
		}
		return nil
	}
}
