package handler
import "fmt"
func PrintEmbedSize() {
	fmt.Printf("Index size: %d, Admin size: %d\n", len(webUITemplate), len(adminUITemplate))
}
