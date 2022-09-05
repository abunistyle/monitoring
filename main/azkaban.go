package main

import (
	"monitoring/controllers/bigdata"
)

func main() {
	azkabanMontior := bigdata.AzkabanMotior{}
	azkabanMontior.Sendnotice()

}
