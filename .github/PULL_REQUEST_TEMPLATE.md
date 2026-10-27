## O que muda

-

## Motivo

- Issue relacionada: #

## Como validar

- Comandos executados e saída esperada (por exemplo `mu notes search foo --limit 5`):

## Contrato da CLI

- [ ] Saída em texto e em `--json` cobertas por arquivo golden
- [ ] Códigos de saída afirmados nos testes dos comandos
- [ ] Todo comando novo tem `Example` executável e `mu --help` segue abaixo de 40 linhas
- [ ] Nenhum endpoint, campo ou status fora do `openapi.yaml` do servidor
- [ ] Token nunca vem de flag nem aparece em saída, erro ou log

## Checklist

- [ ] PR pequeno e focado em uma mudança
- [ ] Commits atômicos no formato `tipo: descrição`, sem escopo
- [ ] Casos limite considerados (vazio, zero, erro do servidor, conflito)
- [ ] `gofmt -l .`, `go vet ./...`, `golangci-lint run` e `go test ./... -race` passando
