# Regras e Diretrizes para Desenvolvimento do Sistema

## Diretrizes Gerais de Desenvolvimento

### 1. Análise Profunda
- **Antes de qualquer implementação**, realizar análise completa dos requisitos
- **Entender o contexto de domínio** antes de escrever código
- **Identificar dependências** entre componentes
- **Considerar escalabilidade** e manutenibilidade desde o início

### 2. Estrutura e Organização
- **Seguir estritamente a arquitetura Clean Architecture** definida
- **Manter separação clara de responsabilidades** entre camadas
- **Organizar código por domínios** conforme estrutura de pastas
- **Evitar acoplamento entre camadas** superiores e inferiores

### 3. Código e Documentação
- **Não utilizar emojis** em nenhum código, comentário ou documentação
- **Manter nomes de variáveis, funções e tipos em inglês**
- **Escrever documentação técnica em português** para facilitar entendimento da equipe
- **Comentar apenas quando necessário** para explicar lógica complexa
- **Seguir padrões Go** para formatação e estilo de código

### 4. Trabalho Modular
- **Dividir funcionalidades em módulos independentes**
- **Implementar um domínio por vez** de forma completa
- **Evitar implementações parciais** que deixem funcionalidades quebradas
- **Garantir coesão interna** em cada módulo
- **Sempre pessa minha autorização para mexer no código**

## Diretrizes Técnicas Específicas

### 1. Domínio e Entidades
- **Entidades devem representar conceitos de negócio reais**
- **Propriedades das entidades devem ser bem definidas**
- **Incluir validações de domínio** nas entidades quando apropriado
- **Utilizar UUIDs** para identificadores conforme definido no schema do banco

### 2. Repositórios e Persistência
- **Definir interfaces de repositório no domínio**
- **Implementar repositórios concretos na infraestrutura**
- **Utilizar SQL raw** em vez de ORMs conforme decisão arquitetural
- **Utilizar sqlx** para facilitar operações comuns com SQL

### 3. Serviços de Aplicação
- **Implementar casos de uso nos serviços de aplicação**
- **Manter serviços de aplicação agnósticos de framework**
- **Tratar erros de forma apropriada** e retornar tipos consistentes
- **Não expor entidades de domínio diretamente** nas interfaces

### 4. Interfaces HTTP
- **Utilizar Gin Framework** conforme decisão tecnológica
- **Separar handlers, middleware e validadores** claramente
- **Utilizar DTOs** para entrada e saída de dados na API
- **Implementar validação de entrada** com validator.v9
- **Seguir padrões RESTful** para endpoints

### 5. Autenticação e Autorização
- **Utilizar JWT** conforme decisão arquitetural
- **Implementar middleware de autenticação** reutilizável
- **Verificar permissões** de forma centralizada
- **Não armazenar informações sensíveis** em tokens JWT

### 6. Tratamento de Erros
- **Definir tipos de erro customizados** para o domínio
- **Mapear erros internos** para respostas HTTP apropriadas
- **Não expor detalhes de implementação** em respostas de erro
- **Logar erros com contexto suficiente** para debugging

### 7. Logging
- **Utilizar Zap** conforme decisão tecnológica
- **Incluir request IDs** para rastreabilidade
- **Logar informações contextuais** mas evitar dados sensíveis
- **Manter estrutura consistente** nos logs

### 8. Configuração
- **Utilizar variáveis de ambiente** para configuração
- **Definir valores padrão sensíveis** para configurações
- **Validar configurações no startup** da aplicação
- **Não hardcodear** valores de configuração

## Diretrizes de Qualidade e Testes

### 1. Testes Estratégicos
- **Focar em testes de unidade** para lógica de domínio complexa
- **Priorizar testes de integração** para fluxos críticos
- **Implementar testes E2E** para principais funcionalidades de negócio
- **Evitar testes redundantes** ou triviais

### 2. Código de Qualidade
- **Seguir princípios SOLID** quando apropriado
- **Manter funções e métodos pequenos** e com responsabilidade única
- **Evitar duplicação de código** através de abstrações adequadas
- **Utilizar injeção de dependência** para facilitar testes

### 3. Revisão e Refatoração
- **Revisar código implementado** antes de considerar completo
- **Refatorar quando necessário** para manter qualidade
- **Manter cobertura de testes** durante refatorações
- **Documentar decisões de design** importantes

## Diretrizes de Trabalho com Agentes de IA

### 1. Otimização de Tokens
- **Ser direto e específico** nas instruções para agentes
- **Evitar repetições** desnecessárias de contexto
- **Fornecer exemplos** quando útil para reduzir tokens gastos
- **Dividir tarefas complexas** em subtarefas menores

### 2. Uso de Ferramentas MCP
- **Utilizar context7** para consulta à documentação quando necessário
- **Utilizar postgres** para esclarecimentos sobre schema do banco
- **Evitar uso excessivo** que consuma tokens desnecessariamente
- **Especificar claramente** o que se espera como resultado

### 3. Implementação Eficiente
- **Focar em implementar funcionalidade completa** por vez
- **Evitar implementações parciais** que não agreguem valor imediato
- **Testar código implementado** antes de prosseguir
- **Corrigir problemas imediatamente** em vez de acumular débito técnico

### 4. Controle de Versão
- **Criar branch por feature** a partir da main
- **Commits pequenos e com mensagens claras**
- **Pull requests com descrição completa** do que foi implementado
- **Revisão de código** antes do merge para main

## Diretrizes de Performance e Segurança

### 1. Performance
- **Otimizar consultas ao banco** com índices apropriados
- **Utilizar cache estrategicamente** para dados frequentemente acessados
- **Evitar N+1 queries** em relacionamentos
- **Monitorar uso de memória** e CPU durante desenvolvimento

### 2. Segurança
- **Validar todas as entradas** de usuário
- **Sanitizar dados** antes de persistir ou exibir
- **Utilizar prepared statements** para evitar SQL injection
- **Proteger endpoints sensíveis** com autenticação e autorização
- **Não logar informações sensíveis** como senhas ou tokens

### 3. Multi-tenancy
- **Isolar dados de tenants** em todas as operações
- **Verificar contexto de tenant** em todas as requisições
- **Evitar vazamento de dados** entre tenants
- **Implementar políticas de acesso** baseadas em tenant

## Diretrizes de Documentação

### 1. Documentação Técnica
- **Manter documentação atualizada** conforme implementação
- **Documentar decisões de arquitetura** importantes
- **Incluir exemplos de uso** para APIs e componentes complexos
- **Utilizar formato markdown** para consistência

### 2. Documentação da API
- **Utilizar Swagger/OpenAPI** para documentação automática
- **Incluir exemplos de requisições e respostas**
- **Documentar códigos de erro** possíveis
- **Manter documentação sincronizada** com implementação

### 3. Comentários no Código
- **Comentar apenas quando necessário** para entender lógica complexa
- **Evitar comentários óbvios** ou redundantes
- **Atualizar comentários** quando modificar código
- **Não utilizar emojis** ou formatação especial nos comentários

## Diretrizes de Manutenção

### 1. Código Legível
- **Utilizar nomes descritivos** para variáveis, funções e tipos
- **Manter estrutura consistente** em todo o código
- **Evitar abreviações** que não sejam comumente entendidas
- **Seguir convenções da comunidade Go**

### 2. Evolução do Sistema
- **Planejar para extensibilidade** desde o início
- **Manter retrocompatibilidade** quando possível
- **Documentar breaking changes** claramente
- **Migrar dados gradualmente** em atualizações

### 3. Monitoramento
- **Implementar logging adequado** para debugging
- **Adicionar métricas relevantes** para monitoramento
- **Configurar alertas** para condições críticas
- **Manter traces** para diagnóstico de problemas