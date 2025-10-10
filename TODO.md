# TODO

## Investigação API Homolog (AWS) – 2025-10-10 08:07–08:20 UTC

### Problemas Observados
- Rotação contínua de tasks ECS (`3f75…`, `714b…`, `1ef1…`) ocupando o serviço em pares; tasks novas sobem, passam health-check e logo são desligadas enquanto a task antiga ainda mantém o lock Redis do `client_registry`.
- Eventos `receipt` e posteriormente `message` para a instância `620ac57d-5b8b-47cd-aa36-98fea4ae7a2e` excedendo o deadline do pipeline (`context deadline exceeded` no `event_integration`), com alertas `Node handling is taking long` chegando a 2m30s.
- Chamadas intermitentes do webhook externo `https://hook.prod.setupautomatizado.com.br/...` completando com sucesso, mas algumas permanecendo travadas até o limite de 150s, gerando o timeout citado acima.
- Timeouts recorrentes em `contact_metadata_cache` ao buscar fotos de perfil (`info query timed out` e `websocket not connected`), indicando congestão/instabilidade da conexão ao WhatsApp durante os atrasos.
- Durante o shutdown da task antiga (08:18:57Z), ainda houve entrega de eventos; o orquestrador já havia sido desmontado, resultando em `no handler registered` e falha ao enviar recibo (`websocket not connected`), com risco de perda de eventos nesse intervalo.

### Ações Recomendadas
1. **Analisar latência do webhook externo**: coletar métricas de tempo de resposta do endpoint `hook.prod.setupautomatizado.com.br` e confirmar se há throttling, gargalos de rede ou problemas no consumidor final que mantenham conexões abertas por >150s; avaliar retries assíncronos/batch menor.
2. **Revisar tempo de shutdown/drain**: garantir que a sequência de desligamento do serviço aguarde o escoamento do `event_integration` e do `event_buffer` antes de desmontar o orquestrador e desconectar o cliente WhatsApp, evitando eventos perdidos.
3. **Monitorar timeouts de metadata**: após normalizar o item (1), conferir se os `info query timed out` desaparecem; caso persistam, investigar latência/jitter na conexão com os servidores do WhatsApp e possíveis limites de concorrência.
4. **Validar motivo da rotação de tasks**: confirmar com o pipeline/infra se a rotação foi um deploy esperado ou reinícios automáticos por saúde/CPU; ajustar janelas ou número de tasks para reduzir corridas pelo lock Redis e minimizar impacto durante trocas.
5. **Adicionar métricas/alertas específicos**: instrumentar o `event_integration` com histogramas para latência e contadores de timeout, e configurar alertas no observability stack para disparar antes de atingir o limite de 150s.

