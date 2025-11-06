package main

import (
	"fmt"
	"bytes"
	"os"
	"time"
	"net/http"
	"encoding/json"


	"github.com/go-faker/faker/v4"
)

type Subs struct {
	Email    string `faker:"email"`
}