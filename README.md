
# GitHub Last Commit Author Fetcher

Este utilitário em Go permite ler um arquivo CSV com `repository_id` de repositórios do GitHub e gerar um novo arquivo CSV contendo o nome e e-mail do autor do **último commit** na branch principal de cada repositório.

---

## 🧰 Requisitos

- Go 1.18 ou superior
- Um token de acesso pessoal (PAT) do GitHub com permissão `public_repo` (ou `repo` se incluir repositórios privados)
- Um arquivo `repos.csv` com a seguinte estrutura:

```
repository_id
10270250
1296269
```

---

## 🚀 Como usar

1. **Clone o repositório e entre no diretório:**

```bash
git clone https://github.com/bquerino/github-fetch-data.git
cd gh-author
```

2. **Inicialize o projeto (caso ainda não esteja):**

```bash
go mod init github.com/bquerino/github-fetch-data
```

3. **Edite `main.go` e defina seu token GitHub:**

Substitua o valor da constante `GITHUB_TOKEN` pelo seu token válido.

4. **Rode o programa:**

```bash
go run main.go
```

5. **Resultado**

Um arquivo `resultado.csv` será gerado no mesmo diretório com os campos:

```
repository_id,repository_name,author_name,author_email
```

---

## 🗂️ Estrutura do Projeto

```
gh-author/
├── main.go         # Código principal
├── repos.csv       # Arquivo de entrada com repository_id
├── resultado.csv   # Arquivo de saída com os dados do commit
├── go.mod
```

---

## 🧪 Exemplo

Com um `repos.csv` contendo:

```
repository_id
10270250
1296269
```

O `resultado.csv` poderá conter:

```
repository_id,repository_name,author_name,author_email
10270250,facebook/react,gaearon,gaearon@users.noreply.github.com
1296269,octocat/Hello-World,octocat,octocat@github.com
```

---

## 📄 Licença

Este projeto está sob a licença MIT.
