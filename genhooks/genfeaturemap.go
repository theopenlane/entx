package genhooks

import (
	"html/template"
	"log"
	"os"
	"path/filepath"
	"slices"
	"sort"

	"entgo.io/ent/entc/gen"

	"github.com/theopenlane/entx"
)

type featureItem struct {
	Type     string
	Features []string
}

type featureMap struct {
	Items []featureItem
}

// GenFeatureMap generates a Go file that maps ent types to features
func GenFeatureMap(outputDir string) gen.Hook {
	return func(next gen.Generator) gen.Generator {
		return gen.GenerateFunc(func(g *gen.Graph) error {
			var items []featureItem
			ant := &entx.FeatureAnnotation{}
			for _, node := range g.Nodes {
				if raw, ok := node.Annotations[ant.Name()]; ok {
					if err := ant.Decode(raw); err == nil {
						// copy slice
						feats := slices.Clone(ant.Features)
						items = append(items, featureItem{Type: node.Name, Features: feats})
					}
				}
			}

			sort.Slice(items, func(i, j int) bool { return items[i].Type < items[j].Type })

			tmpl := createFeatureTemplate()

			if err := os.MkdirAll(outputDir, 0o755); err != nil {
				return err
			}

			filePath := filepath.Join(outputDir, "features.go")

			file, err := os.Create(filePath)
			if err != nil {
				log.Fatalf("Unable to create file: %v", err)
			}

			defer file.Close()

			data := featureMap{Items: items}
			if err := tmpl.Execute(file, data); err != nil {
				log.Fatalf("Unable to execute template: %v", err)
			}

			return next.Generate(g)
		})
	}
}

// createFeatureTemplate creates a template for generating the feature map
func createFeatureTemplate() *template.Template {
	tmpl, err := template.New("features.tmpl").ParseFS(_templates, "templates/features.tmpl")
	if err != nil {
		log.Fatalf("Unable to parse template: %v", err)
	}
	return tmpl
}
