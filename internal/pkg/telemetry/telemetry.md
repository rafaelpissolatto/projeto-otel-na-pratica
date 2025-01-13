# Implementação de Telemetria

Este projeto implementa telemetria usando OpenTelemetry para fornecer capacidades de rastreamento, métricas e logs. A configuração de telemetria é projetada para ajudar a monitorar e depurar a aplicação coletando e exportando dados de telemetria.

## Inicialização

A configuração de telemetria é inicializada na função `InitTelemetry`, que configura rastreamentos, métricas e logs.

### Traces (rastreamentos)

O rastreamento é inicializado na função `InitTraces` localizada em [internal/pkg/telemetry/traces.go](internal/pkg/telemetry/traces.go). Esta função configura um exportador de rastreamento OTLP e um provedor de rastreamento.

```go
func InitTraces(ctx context.Context) error {
    exp, err := otlptracehttp.New(ctx, otlptracehttp.WithInsecure())
    if err != nil {
        return err
    }

    provider := trace.NewTracerProvider(trace.WithBatcher(exp))
    otel.SetTracerProvider(trace.NewTracerProvider(trace.WithBatcher(exp)))
    go func() {
        <-ctx.Done()
        if err := provider.Shutdown(ctx); err != nil {
            log.Printf("Error shutting down TracingProvider: %v", err)
        }
    }()
    return nil
}
```


### Metrics
As métricas são inicializadas na função InitMetrics localizada em metrics.go. Esta função configura um exportador de métricas OTLP e um provedor de medição.

```go
func InitMetrics(ctx context.Context) error {
    exp, err := otlpmetrichttp.New(ctx, otlpmetrichttp.WithInsecure())
    if err != nil {
        return err
    }

    provider := metric.NewMeterProvider(metric.WithReader(metric.NewPeriodicReader(exp)))
    otel.SetMeterProvider(provider)

    go func() {
        <-ctx.Done()
        if err := provider.Shutdown(ctx); err != nil {
            log.Printf("Error shutting down MeterProvider: %v", err)
        }
    }()

    return nil
}
```

### Logs
A configuração de logs é inicializada na função InitLogs localizada em logs.go. Esta função configura um logger para a aplicação.

```go
func InitLogs(ctx context.Context) error {
    // por favor olhe o codigo aplicado
}
```

## Uso
Para inicializar a telemetria na sua aplicação, chame a função InitTelemetry com um contexto:

# cmd/all-in-one, etc
```go
func main() {
    ctx := context.Background()
    telemetry.InitTelemetry(ctx)
    // por favor olhe o codigo aplicado
}
```

## Issues

Ao iniciar o projeto através do all-in-one localmente, recemos o erro:

```shell
$ go run main.go 

panic: nats: no servers available for connection

goroutine 1 [running]:
main.main()
        /home/rafael/rafael_ubuntu/development/github/rafaelpissolatto/projeto-otel-na-pratica/cmd/all-in-one/main.go:53 +0x3b2
exit status 2
```

Para correção, foi adicionado um docker-compose.yaml para rodar o `NATS` através de container:

```yaml
services:
  nats:
    image: nats:alpine
    container_name: nats
    restart: always
    command: -c /etc/nats/nats.conf
    ports:
      - "4222:4222"
      - "6222:6222"
      - "8222:8222"
    volumes:
      - ./configs/nats/nats.conf:/etc/nats/nats.conf
    networks:
      - local

```
Note: Por favor, check o arquivo de configuração reference ao NATS em `configs/nats/nats.conf`

E também para poder iniciar a stream à partir da configuração do próprio código/app:

```go
	_, err = js.CreateOrUpdateStream(ctx, jetstream.StreamConfig{
		Name:     cfg.NATS.Stream,
		Subjects: []string{cfg.NATS.Subject},
	})
	if err != nil {
		return nil, err
	}
```
Note: `app/payment.go:#00`


# //TODO: add screenshots