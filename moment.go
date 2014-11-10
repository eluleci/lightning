package main

type Moment struct {

	CreatedAt         string `json:"createdAt"`
	Image             []string `json:"images"`
	Type              string `json:"typeName"`
	Value             string `json:"typeValue"`
}

