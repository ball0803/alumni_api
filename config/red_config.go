package config

type RouteREDConfig struct {
	Priority     int
	MinThreshold int
	MaxRequests  int
}

var REDRouteConfigs = map[string]RouteREDConfig{
	"/v1/auth/login": {
		MinThreshold: 2,
		MaxRequests:  3,
	},
	"/v1/auth/request_OTR": {
		MinThreshold: 2,
		MaxRequests:  5,
	},
	"/v1/auth/register/user": {
		MinThreshold: 2,
		MaxRequests:  5,
	},
	"/v1/auth/register/alumnus": {
		MinThreshold: 2,
		MaxRequests:  5,
	},
	"/v1/posts": {
		MinThreshold: 80,
		MaxRequests:  100,
	},
	"/v1/users": {
		MinThreshold: 50,
		MaxRequests:  75,
	},
}
