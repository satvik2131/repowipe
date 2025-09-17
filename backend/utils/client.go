package utils

import "github.com/go-resty/resty/v2"


var Client *resty.Client = resty.New().
SetHeader("Accept", "application/vnd.github+json")
