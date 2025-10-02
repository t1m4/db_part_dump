package dump

import (
	"fmt"
	"log"
	"reflect"
	"strings"
)

func buildStringFromSet(fkIds map[string]bool) string {
	result := make([]string, 0)
	for fkId := range fkIds {
		result = append(result, fmt.Sprintf("%v", fkId))
	}
	return strings.Join(result, ", ")
}

func AnyToString(value any) string {
	switch v := value.(type) {
	case []byte:
		return fmt.Sprintf("'%s'", string(v))
	case int, int32, int64:
		return fmt.Sprintf("%v", v)
	default:
		log.Fatalf("Wrong type for fk id. Does not support %s", reflect.TypeOf(v))
	}
	return ""
}
