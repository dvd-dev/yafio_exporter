package main

import (
	"fmt"
	"go/types"
	"log"
	"regexp"
	"strings"

	"golang.org/x/tools/go/packages"
)

var matchFirstCapRex = regexp.MustCompile("(.)([A-Z][a-z]+)")
var matchAllCapRex = regexp.MustCompile("([a-z0-9])([A-Z])")

func loadStruct(pkg *packages.Package, name string) *types.Struct {
	obj := pkg.Types.Scope().Lookup(name)
	if obj == nil {
		log.Fatalf("%s could not find struct", name)
	}
	st, ok := obj.Type().Underlying().(*types.Struct)
	if !ok {
		log.Fatalf("%s is not a struct", name)
	}
	return st
}
func loadPackage(name string) (*packages.Package, error) {
	pkgs, err := packages.Load(
		&packages.Config{
			Mode: packages.NeedName | packages.NeedImports | packages.NeedDeps |
				packages.NeedTypes | packages.NeedTypesInfo | packages.NeedSyntax}, ".")
	if err != nil {
		log.Fatalf("failed to load package: %s", err)
	}
	if len(pkgs) != 1 {
		return nil, fmt.Errorf("expected 1 package, got %d", len(pkgs))
	}
	return pkgs[0], nil
}
func fieldByName(name string, s *types.Struct) (*types.Var, error) {
	for i := 0; i < s.NumFields(); i++ {
		field := s.Field(i)
		if field.Name() == name {
			return field, nil
		}
	}
	return nil, fmt.Errorf("field %s not found", name)
}

func ToSnakeCase(str string) string {
	if len(str) == 0 {
		return str
	}
	str = fmt.Sprintf("%s%s", strings.ToLower(string(str[0])), str[1:])
	snake := matchFirstCapRex.ReplaceAllString(str, "${1}_${2}")
	snake = matchAllCapRex.ReplaceAllString(snake, "${1}_${2}")
	return strings.ToLower(snake)
}
func inMap(key string, list []string) bool {
	for _, b := range list {
		if b == key {
			return true
		}
	}
	return false
}
