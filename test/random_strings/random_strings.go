package random_strings

import (
	"github.com/brunvieira/lotus"
	"math/rand"
	"strings"
	"time"
)

func generateRandomStrings(ctx *lotus.Context) {
	charSet := "abcdedfghijklmnopqrstuvxwyzABCDEFGHIJKLMNOPRSTUVXWYZ"
	rand.Seed(time.Now().Unix())
	var output strings.Builder
	length := 10
	for i := 0; i < length; i++ {
		random := rand.Intn(len(charSet))
		randomChar := charSet[random]
		output.WriteString(string(randomChar))
	}
	ctx.Response.Header.Set("content-type", "text/plain")
	ctx.WriteString(output.String())
}
