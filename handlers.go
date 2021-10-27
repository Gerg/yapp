package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"gopkg.in/yaml.v2"
)

type handledRoute struct {
	method string
	path   string
	config routeConfig
}

func setUpHandlers(config appConfig, yttPath string) {
	for _, route := range getHandledRoutes(config.Routes) {
		route := route // lol

		fmt.Printf("Setting up route %s %s...\n", route.method, route.path)

		http.HandleFunc(route.path, func(res http.ResponseWriter, req *http.Request) {
			if req.Method == route.method {
				res.Header().Set("Content-Type", "text/x-yaml")

				fmt.Printf("%s %s\n", req.Method, route.path)

				if req.Method == "POST" {
					reqBodyBytes, _ := ioutil.ReadAll(req.Body)
					reqBody := make(map[string]interface{})
					_ = yaml.Unmarshal(reqBodyBytes, reqBody)

					err := runYTT(yttPath, &config, map[string]interface{}{
						"request": reqBody,
					})
					if err != nil {
						fmt.Printf("error running ytt: %v", err)
						http.Error(res, "uh oh", 500)
						return
					}

					// replace route config with new config generated by ytt
					route.config = config.Routes[fmt.Sprintf("%s %s", route.method, route.path)]
				}

				if route.config.Status != nil {
					res.WriteHeader(*route.config.Status)
				}

				if route.config.Body != nil {
					responseBody, err := yaml.Marshal(route.config.Body)
					if err != nil {
						fmt.Printf("error marshalling response body: %v", err)
						http.Error(res, "uh oh", 500)
						return
					}

					res.Write(responseBody)
				}

				return
			}

			res.WriteHeader(404)
		})
	}
}

func getHandledRoutes(routes map[string]routeConfig) []handledRoute {
	var handledRoutes []handledRoute

	for routeWithMethod, routeConfig := range routes {
		parts := strings.Split(routeWithMethod, " ")

		handledRoutes = append(handledRoutes, handledRoute{
			method: parts[0],
			path:   parts[1],
			config: routeConfig,
		})
	}

	return handledRoutes
}
