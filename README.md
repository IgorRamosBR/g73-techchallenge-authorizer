# g73-techchallenge-authorizer

[![Go Reference](https://pkg.go.dev/badge/golang.org/x/example.svg)](https://pkg.go.dev/golang.org/x/example)

Esta aplicação contém os arquivos necessários para configurar um pipeline de implantação de uma função Lambda na AWS. A função Lambda é usada como um authorizer para autenticar usuários com base em seu CPF por meio do API Gateway.


## Tecnologias Utilizadas

- **Go:** A função Lambda é escrita em Go, uma linguagem de programação eficiente e poderosa.
- **Gin:** O framework Gin é usado para criar endpoints RESTful e lidar com as solicitações HTTP.
- **AWS Lambda:** A função Lambda é implantada na AWS e é responsável por autenticar os usuários.
- **API Gateway:** O API Gateway da AWS é usado para expor a função Lambda como uma API RESTful acessível externamente.
- **DynamoDB:** O DynamoDB é usado como o banco de dados NoSQL para armazenar informações do usuário.



## Descrição

A pipeline de implantação é acionada sempre que uma nova tag é enviada para o repositório. Ele realiza as seguintes etapas:

**1. Checkout do Repositório:** Clona o repositório para o ambiente de execução do GitHub Actions.

**2. Configuração do AWS CLI:** Configura as credenciais da AWS necessárias para implantar a função Lambda.

**3. Configuração do Go:** Configura o ambiente com a versão especificada do Go.

**4. Compilação da Função Lambda:** Compila a função Lambda usando as configurações especificadas.

**5. Compactação da Função Lambda:** Compacta a função Lambda compilada em um arquivo zip.

**6. Implantação na AWS Lambda:** Atualiza o código da função Lambda na AWS com o arquivo zip gerado.



## Funcionalidades
- **Autenticação por CPF:** Os usuários podem ser autenticados por meio de seus CPFs. Isso é feito enviando uma solicitação para o endpoint **/authorize** com o CPF do usuário no corpo da solicitação.
- **Criação de Usuários:** Os administradores podem criar novos usuários enviando uma solicitação para o endpoint **/user** com os detalhes do usuário no corpo da solicitação.



## Como Executar
Para executar esta aplicação, siga estas etapas:

**1.** Certifique-se de que as variáveis de ambiente secretas necessárias estão configuradas no GitHub.

**2.** Adicione os arquivos **deploy_function.yml** e **main.go** à raiz do seu repositório.

**3.** Ao criar uma nova tag e fazer o push para o repositório, o pipeline de implantação será acionado automaticamente.


## Detalhes do Código

O arquivo **main.go** contém o código da função Lambda escrita em Go. Ele usa o framework Gin para criar endpoints RESTful para autorizar usuários e criar novos usuários. O código também interage com o DynamoDB da AWS para armazenar e recuperar informações do usuário.

