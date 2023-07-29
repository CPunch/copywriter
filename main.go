package main

import (
	"fmt"
)

func main() {
	post := genBlogPost()
	Success("Generated blog post !")
	fmt.Println(post)
}
