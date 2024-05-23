package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/lclpedro/ddos/pkg/threading"
	"net/http"
	"os"
	"time"

	"github.com/spf13/cobra"
)

type RequestData struct {
	Endpoint    string
	Workers     int
	Concurrency int
}

func makeRequest(ctx context.Context, data *RequestData) error {
	client := &http.Client{
		Timeout: 5 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return nil
		},
	}
	req, err := http.NewRequestWithContext(ctx, "GET", data.Endpoint, nil)
	if err != nil {
		return fmt.Errorf("erro ao criar requisição: %s", err.Error())
	}

	resp, err := client.Do(req)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			fmt.Printf("Status code: %d\n", http.StatusRequestTimeout)
			return fmt.Errorf("tempo limite excedido")
		}
		return fmt.Errorf("erro ao fazer a requisição: %s", err.Error())
	}

	fmt.Printf("Status code: %d\n", resp.StatusCode)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("status code Error: %d", resp.StatusCode)
	}

	return nil
}

func runWorkers(data *RequestData) error {
	startHour := time.Now()
	workers := threading.NewWorkerPool(data.Concurrency)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second) // Tempo limite de 10 segundos
	defer cancel()
	for i := 0; i < data.Workers; i++ {
		workers.RunJob([]any{}, func(_dataset []interface{}) error {
			if err := makeRequest(ctx, data); err != nil {
				return err
			}
			return nil
		})
	}

	workers.Wait()
	endHour := time.Now()
	fmt.Printf("\n\n=================== REPORT EXECUTION =====================\n\n")
	fmt.Println("Tempo de execução:", endHour.Sub(startHour))
	fmt.Println("\nQuantidade de Requests:", workers.NumOfExecutions())
	fmt.Println("Quantidade de Sucesso (200):", workers.NumOfExecutions()-workers.NumOfFailures())
	fmt.Println("Quantidade de Falhas (4XX, 5XX):", workers.NumOfFailures())
	fmt.Println("Ultima mensagem de Erro:", workers.Error())

	return nil
}

func main() {
	var rootCmd = &cobra.Command{
		Use:   "app",
		Short: "Uma aplicação de teste de requisições",
	}

	var endpoint string
	rootCmd.Flags().StringVarP(&endpoint, "url", "u", "", "Endpoint para a requisição")
	rootCmd.MarkFlagRequired("endpoint")

	var workers int
	rootCmd.Flags().IntVarP(&workers, "requests", "r", 1, "Número de workers simultâneos")

	var concurrency int
	rootCmd.Flags().IntVarP(&concurrency, "concurrency", "c", 1, "Número de requests simultâneos")

	rootCmd.Run = func(cmd *cobra.Command, args []string) {
		data := &RequestData{
			Endpoint:    endpoint,
			Workers:     workers,
			Concurrency: concurrency,
		}

		if err := runWorkers(data); err != nil {
			fmt.Println("Erro ao executar os workers:", err)
			os.Exit(1)
		}
	}

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
