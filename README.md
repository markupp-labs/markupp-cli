# markupp-cli

`mu` é o cliente de linha de comando do [Markupp](https://github.com/markupp-labs/markupp). Ele
lê e escreve as notas de um servidor Markupp pela API REST pública (`/v1`), com os mesmos
privilégios de qualquer outro cliente.

O mesmo binário atende dois públicos:

- **Pessoas** usam `mu tui`, uma interface interativa no molde do K9s, e subcomandos para
  tarefas rápidas.
- **Agentes de IA** usam só os subcomandos: texto por padrão, `--json` com o payload da API sem
  alteração e `--ndjson` nas listagens. Não há prompts fora da `mu tui`.

## Instalação

    go install github.com/markupp-labs/markupp-cli/cmd/mu@latest

Os binários estáticos de linux, darwin e windows (amd64 e arm64) saem em cada tag `v*` pelo
goreleaser.

## Primeiros passos

    mu login https://markupp.example.com   # abre o navegador (OAuth 2.1 com PKCE)
    mu status
    mu vaults list
    mu notes search "reunião" --limit 5
    mu notes get ID --content > nota.md
    mu notes edit ID --file nota.md --since 2026-10-26T12:00:00Z
    mu tui

`mu --help` lista todos os comandos e `mu <comando> --help` traz os detalhes e exemplos.

## Configuração

A configuração fica em `$XDG_CONFIG_HOME/markupp/cli.json` (ou `~/.config/markupp/cli.json`),
com permissão `0600`. A precedência é flag, depois variável de ambiente, depois arquivo.

| Variável         | Flag       | Uso                                                        |
|------------------|------------|------------------------------------------------------------|
| `MARKUPP_URL`    | `--url`    | endereço do servidor                                       |
| `MARKUPP_TOKEN`  | nenhuma    | token de acesso pronto, para CI; nunca é renovado          |
| `MARKUPP_TENANT` | `--tenant` | tenant, quando a conta participa de vários                 |
| `MARKUPP_VAULT`  | `--vault`  | cofre padrão dos comandos de notas e de grafo              |

O token nunca vem de flag e nunca é impresso. `http://` só é aceito para localhost, salvo com
`--allow-plain-http`.

## Códigos de saída

| Código | Significado                                              |
|--------|----------------------------------------------------------|
| 0      | sucesso                                                  |
| 1      | falha interna ou resposta malformada                     |
| 2      | uso incorreto ou requisição inválida (400)               |
| 3      | sem servidor configurado ou sem login válido (401)       |
| 4      | acesso negado (403)                                      |
| 5      | não encontrado (404), inclusive endpoint ausente         |
| 6      | conflito (409), por exemplo edição concorrente           |
| 7      | limite de requisições (429)                              |
| 8      | servidor indisponível, erro de rede ou 5xx               |
| 130    | interrompido                                             |

## Desenvolvimento

    make test
    make lint

A versão do Go está fixada em `mise.toml`. As regras do repositório estão em
[AGENTS.md](AGENTS.md).

## Licença

MIT.
