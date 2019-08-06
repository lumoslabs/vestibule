// +build ignore

package main

func main() {
  er := vfsgen.Generate(newAssetsFs("manifests"), vfsgen.Options{
		PackageName:  "data",
		BuildTags:    "!dev",
		VariableName: "Assets",
  })
  if er != nil {
		log.Fatalln(er)
	}
}
