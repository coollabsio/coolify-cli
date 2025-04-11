package openapi

//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest -config oapi-codegen.yaml ../../../openapi.yaml

// We need to implement methods on some structs we do this here so the generated code doesnt break
func (s *Server) GetFilterValue() string {
	return *s.Name
}

func (p *PrivateKey) GetFilterValue() string {
	return *p.Name
}
