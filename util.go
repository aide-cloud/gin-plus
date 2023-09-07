package ginplus

import (
	"fmt"
	"path"
	"reflect"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	GinHandleFunc = "gin.HandlerFunc"
)

// 判断该方法返回值是否为 gin.HandlerFunc类型
func isHandlerFunc(t reflect.Type) bool {
	if t.Kind() != reflect.Func {
		return false
	}
	if t.NumOut() != 1 {
		return false
	}
	return t.Out(0).String() == GinHandleFunc
}

// isCallBack 判断是否为CallBack类型
func isCallBack(t reflect.Type) (reflect.Type, reflect.Type, bool) {
	// 通过反射获取方法的返回值类型
	if t.Kind() != reflect.Func {
		return nil, nil, false
	}

	if t.NumIn() != 3 || t.NumOut() != 2 {
		return nil, nil, false
	}

	if t.Out(1).String() != "error" {
		return nil, nil, false
	}

	if t.In(1).String() != "context.Context" {
		return nil, nil, false
	}

	// new一个out 0的实例和in 2的实例
	req := t.In(2)
	resp := t.Out(0)

	return req, resp, true
}

func (l *GinEngine) genRoute(parentGroup *gin.RouterGroup, controller any, skipAnonymous bool) {
	t := reflect.TypeOf(controller)

	tmp := t
	for tmp.Kind() == reflect.Ptr {
		tmp = t.Elem()
	}

	if !l.isPublic(tmp.Name()) {
		return
	}

	var middlewares []gin.HandlerFunc
	midd, isMidd := isMiddlewarer(controller)
	if isMidd {
		//Middlewares方法返回的是gin.HandlerFunc类型的切片, 中间件
		middlewares = midd.Middlewares()
	}

	basePath := l.routeNamingRuleFunc(tmp.Name())
	ctrl, isCtrl := isController(controller)
	if isCtrl {
		basePath = ctrl.BasePath()
	}
	parentRouteGroup := parentGroup
	if parentRouteGroup == nil {
		parentRouteGroup = l.Group("/")
	}
	routeGroup := parentRouteGroup.Group(path.Join(basePath), middlewares...)

	methoderMiddlewaresMap := make(map[string][]gin.HandlerFunc)
	methoderMidd, privateMiddOk := isMethoderMiddlewarer(controller)
	if privateMiddOk {
		methoderMiddlewaresMap = methoderMidd.MethoderMiddlewares()
	}

	if !skipAnonymous {
		for i := 0; i < t.NumMethod(); i++ {
			metheodName := t.Method(i).Name
			if !l.isPublic(metheodName) {
				continue
			}
			privateMidd := methoderMiddlewaresMap[metheodName]
			route := l.parseRoute(metheodName)
			if route == nil {
				continue
			}
			route.Path = path.Join(route.Path)
			// 接口私有中间件
			route.Handles = append(route.Handles, privateMidd...)

			if isHandlerFunc(t.Method(i).Type) {
				// 具体的action
				handleFunc := t.Method(i).Func.Call([]reflect.Value{reflect.ValueOf(controller)})[0].Interface().(gin.HandlerFunc)
				route.Handles = append(route.Handles, handleFunc)
				routeGroup.Handle(strings.ToUpper(route.HttpMethod), route.Path, route.Handles...)
				continue
			}

			// 判断是否为CallBack类型
			req, resp, isCb := isCallBack(t.Method(i).Type)
			if isCb {
				reqName := req.Name()
				respName := resp.Name()
				reqTagInfo := getTag(req)
				apiRoute := ApiRoute{
					Path:       route.Path,
					HttpMethod: strings.ToLower(route.HttpMethod),
					MethodName: metheodName,
					ReqParams: Field{
						Name: reqName,
						Info: reqTagInfo,
					},
					RespParams: Field{
						Name: respName,
						Info: getTag(resp),
					},
				}

				// 处理Uri参数
				for _, tagInfo := range reqTagInfo {
					uriKey := tagInfo.Tags.UriKey
					if uriKey != "" && uriKey != "-" {
						route.Path = path.Join(route.Path, fmt.Sprintf(":%s", uriKey))
					}
				}
				apiRoute.Path = route.Path

				if _, ok := l.apiRoutes[route.Path]; !ok {
					l.apiRoutes[route.Path] = make([]ApiRoute, 0, 1)
				}
				l.apiRoutes[route.Path] = append(l.apiRoutes[route.Path], apiRoute)

				// 具体的action
				handleFunc := l.defaultHandler(controller, t.Method(i), req)
				route.Handles = append(route.Handles, handleFunc)
				routeGroup.Handle(strings.ToUpper(route.HttpMethod), route.Path, route.Handles...)
				continue
			}
		}
	}

	if isStruct(tmp) {
		// 递归获取内部的controller
		for i := 0; i < tmp.NumField(); i++ {
			field := tmp.Field(i)
			for field.Type.Kind() == reflect.Ptr {
				field.Type = field.Type.Elem()
			}
			if !isStruct(field.Type) {
				continue
			}

			if !l.isPublic(field.Name) {
				continue
			}

			// new一个新的controller
			newController := reflect.New(field.Type).Interface()
			l.genRoute(routeGroup, newController, field.Anonymous)
		}
	}
}

// isPublic 判断是否为公共方法
func (l *GinEngine) isPublic(name string) bool {
	if len(name) == 0 {
		return false
	}

	first := name[0]
	if first < 'A' || first > 'Z' {
		return false
	}

	return true
}

// parseRoute 从方法名称中解析出路由和请求方式
func (l *GinEngine) parseRoute(methodName string) *Route {
	method := ""
	routePath := ""

	for prefix, httpMethodKey := range l.httpMethodPrefixes {
		if strings.HasPrefix(methodName, prefix) {
			method = strings.ToLower(httpMethodKey.key)
			p := strings.TrimPrefix(methodName, prefix)
			if p != "" {
				routePath = strings.ToLower(p)
			}
			break
		}
	}

	if method == "" || routePath == "" {
		return nil
	}

	return &Route{
		Path:       "/" + l.routeNamingRuleFunc(routePath),
		HttpMethod: method,
	}
}

// routeToCamel 将路由转换为驼峰命名
func routeToCamel(route string) string {
	if route == "" {
		return ""
	}

	// 首字母小写
	if route[0] >= 'A' && route[0] <= 'Z' {
		route = string(route[0]+32) + route[1:]
	}

	return route
}

// isMiddlewarer 判断是否为Controller类型
func isMiddlewarer(c any) (Middlewarer, bool) {
	midd, ok := c.(Middlewarer)
	return midd, ok
}

// isController 判断是否为Controller类型
func isController(c any) (Controller, bool) {
	ctrl, ok := c.(Controller)
	return ctrl, ok
}

// isMethoderMiddlewarer 判断是否为MethoderMiddlewarer类型
func isMethoderMiddlewarer(c any) (MethoderMiddlewarer, bool) {
	midd, ok := c.(MethoderMiddlewarer)
	return midd, ok
}

// isStruct 判断是否为struct类型
func isStruct(t reflect.Type) bool {
	return t.Kind() == reflect.Struct
}
