package ginplus

import (
	"reflect"
)

type (
	Field struct {
		Name string
		Info []FieldInfo
	}

	FieldInfo struct {
		Type      string
		Name      string
		Tags      Tag
		ChildType string
		Info      []FieldInfo
	}

	Tag struct {
		FormKey string
		UriKey  string
		JsonKey string
		Title   string
		Format  string
		Desc    string
	}
)

var tags = []string{"form", "uri", "json", "title", "format", "desc"}

// 获取结构体tag
func getTag(t reflect.Type) []FieldInfo {
	tmp := t
	for tmp.Kind() == reflect.Ptr {
		tmp = tmp.Elem()
	}

	if tmp.Kind() == reflect.Slice {
		tmp = tmp.Elem()
		for tmp.Kind() == reflect.Ptr {
			tmp = tmp.Elem()
		}
	}

	if tmp.Kind() != reflect.Struct {
		return nil
	}

	fieldList := make([]FieldInfo, 0, tmp.NumField())
	for i := 0; i < tmp.NumField(); i++ {
		field := tmp.Field(i)
		fieldName := field.Name
		fieldType := field.Type.String()
		tagInfo := Tag{
			Title: fieldName,
		}
		for _, tagKey := range tags {
			tagVal, ok := field.Tag.Lookup(tagKey)
			if !ok {
				continue
			}

			switch tagKey {
			case "form":
				tagInfo.FormKey = tagVal
			case "uri":
				tagInfo.UriKey = tagVal
			case "title":
				tagInfo.Title = tagVal
			case "format":
				tagInfo.Format = tagVal
			case "desc":
				tagInfo.Desc = tagVal
			default:
				tagInfo.JsonKey = tagVal
			}
		}

		childType := field.Type
		if childType.Kind() == reflect.Slice {
			childType = childType.Elem()
			for childType.Kind() == reflect.Ptr {
				childType = childType.Elem()
			}
		}

		fieldList = append(fieldList, FieldInfo{
			Type:      fieldType,
			Name:      fieldName,
			Tags:      tagInfo,
			ChildType: childType.String(),
			Info:      getTag(childType),
		})
	}

	return fieldList
}
