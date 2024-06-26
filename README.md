# Teal

Introducing Teal: ETL with Power and Simplicity
In the ever-evolving landscape of data engineering, the need for robust, scalable, and user-friendly tools is paramount. Today, we are thrilled to introduce Teal, a groundbreaking open-source ETL tool designed to redefine your data transformation and orchestration experience.

Teal combines the best elements of industry-leading tools like dbt, Dagster, and Airflow, while addressing the common challenges and limitations associated with traditional Python-based solutions. Our mission is to empower data engineers and analysts with a powerful yet intuitive platform that streamlines complex workflows and enhances productivity.

Why Teal?

- **Scalable Architecture:** Seamlessly scale your data pipelines to handle datasets of any size, ensuring performance and reliability for your growing needs.
Flexible Integration: Effortlessly integrate with a wide range of data sources and destinations, providing unparalleled flexibility and connectivity.
- **Optimized Performance with Go:** TEAL leverages the power of Go's concurrency model, utilizing goroutines and channels to maximize performance and efficiency. This ensures that your data pipelines run swiftly and reliably, making the most of your system's resources.
- **Go Stack Advantage:** Built on the robust and efficient Go stack, TEAL offers high performance, low latency, and excellent scalability. The simplicity and power of Go provide a solid foundation for creating and managing complex ETL workflows.

## QuickStart

### Installation

```bash
go install github.com/go-teal/teal/cmd/teal@latest
```

### Creating your project

```bash
mkdir my_test_project
cd my_test_project
```

### Init you project from scratch

```bash
teal init
```

```bash
❯ ls -al
total 16
drwxr-xr-x@ 6 wwtlf  staff  192 24 Jun 21:23 .
drwxr-xr-x  5 wwtlf  staff  160 24 Jun 21:21 ..
drwxr-xr-x@ 3 wwtlf  staff   96 24 Jun 07:46 assets
-rw-r--r--@ 1 wwtlf  staff  302 24 Jun 07:51 config.yaml
drwxr-xr-x@ 2 wwtlf  staff   64 24 Jun 20:03 docs
-rw-r--r--@ 1 wwtlf  staff  137 24 Jun 07:46 profile.yaml
```

### Update **config.yaml**

```yaml
version: '1.0.0'
module: github.com/my_user/my_test_project
connections:
  - name: default
    type: duckdb
    config:
      path: ./store/test.duckdb            
      extensions:
        - postgres
        - httpfs         
      # extraParams: 
      #   - name: "name"
      #     value: "value"
```

1. `module` param will be used as a module in go.mod
2. Make sure the dir from the `path` exists.

### Update **profile.yaml**

```yaml
version: '1.0.0'
name: 'my-test-project'
connection: 'default'
models: 
  stages:
    - name: staging
    - name: dds  
    - name: mart
```

1. `name` will be used as a name for the binary file

### Generate go project

```bash
teal gen
```

You'll see the following outpout

```bash
project-path: .
config-file: ./config.yaml
Building: staging.addresses.sql
Building: staging.transactions.sql
Building: staging.wallets.sql
Building: dds.dim_addresses.sql
Building: dds.fact_transactions.sql
Building: mart.mart_wallet_report.sql
Files 10
./cmd/my-test-project/main._go .................................................. [OK]
./go.mod ........................................................................ [OK]
./internal/assets/staging.addresses.go .......................................... [OK]
./internal/assets/staging.transactions.go ....................................... [OK]
./internal/assets/staging.wallets.go ............................................ [OK]
./internal/assets/dds.dim_addresses.go .......................................... [OK]
./internal/assets/dds.fact_transactions.go ...................................... [OK]
./internal/assets/mart.mart_wallet_report.go .................................... [OK]
./internal/assets/configs.go .................................................... [OK]
./docs/graph.wsd ................................................................ [OK]
```

1. Rename `main._go` to `my-test-project.go`
2. Uncomment the following line: `_ "github.com/marcboeker/go-duckdb"` in `my-test-project.go`.
3. Run `go mod tidy`
4. Final project structure:

```bash
.
├── assets
│   └── models
│       ├── dds
│       │   ├── dim_addresses.sql
│       │   └── fact_transactions.sql
│       ├── mart
│       │   └── mart_wallet_report.sql
│       └── staging
│           ├── addresses.sql
│           ├── transactions.sql
│           └── wallets.sql
├── cmd
│   └── my-test-project
│       └── main.go
├── config.yaml
├── docs
│   └── graph.wsd
├── go.mod
├── go.sum
├── internal
│   └── assets
│       ├── configs.go
│       ├── dds.dim_addresses.go
│       ├── dds.fact_transactions.go
│       ├── mart.mart_wallet_report.go
│       ├── staging.addresses.go
│       ├── staging.transactions.go
│       └── staging.wallets.go
├── profile.yaml
└── store
    ├── addresses.csv    
    ├── transactions.csv
    └── wallets.csv
```

### Run your project

```bash
go run ./cmd/my-test-project
```

### Explore my-test-project.go

```go
package main

import (
	_ "github.com/marcboeker/go-duckdb"

	"fmt"
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/go-teal/teal/pkg/core"
	"github.com/go-teal/teal/pkg/dags"
	"github.com/my_user/my_test_project/internal/assets"
)

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	fmt.Println("my-test-project")
	core.GetInstance().Init("config.yaml", ".")
	config := core.GetInstance().Config
	dag := dags.InitChannelDag(assets.DAG, assets.PorjectAssets, config, "instance 1")
	wg := dag.Run()
	result := <-dag.Push("TEST", nil, make(chan map[string]interface{}))
	fmt.Println(result)
	dag.Stop()
	wg.Wait()
}
```

What this code does:

1. `dag.Run()` builds DAGs based on Ref from your .sql models, where each node is an asset, and each edge is a GO channel.
2. `result := <-dag.Push("TEST", nil, make(chan map[string]interface{}))` triggers the execution of this DAG synchronously.
3. `dag.Stop()` sends the deactivation command

## Configuration

### config.yaml

```yaml
version: '1.0.0'
module: github.com/my_user/my_test_project
connections:
  - name: default
    type: duckdb
    config:
      path: ./store/test.duckdb            
      extensions:
        - postgres
        - httpfs         
      # extraParams: 
      #   - name: "name"
      #     value: "value"
```

1. Teal supports multiple connections.
2. The following databases are supported at the moment (v0.1.0):
    - [DuckDB](#duckdb), see the specific config params

|Param|Type|Description|
|-----|----|-----------|
|version|String constant|1.0.0|
|module|String|Generated go module name|
|connections|String|Array of database connections|
|connections.name|String|Name of the connection for model profile|
|connections.type|String|Driver name of the database connection, DuckDB, PostgreSQL, etc.|

### profile.yaml

```yaml
version: '1.0.0'
name: 'my-test-project'
connection: 'default'
models: 
  stages:
    - name: staging
    - name: dds  
    - name: mart
```

|Param|Type|Description|
|-----|----|-----------|
|version|String constant|1.0.0|
|name|String|Generated folder name for main.go file|
|connection|String|Connection from `config.yaml` by default|
|models.stages:|Array of stages|list of stages for models. For each stage a folder `assets/models`/`<stage name>` must be created in advance|

## Databases

<a name="duckdb"></a>

### DuckDB

1. To enable DuckDB support, the following line ```_ "github.com/marcboeker/go-duckdb"``` must be added to the ```main.go``` file
2. Specific config params:

|Param|Type|Description|
|-----|----|-----------|
|extensions|Array of strings|List of [DuckDB extenstions](https://duckdb.org/docs/extensions/overview.html). Extenstions will be install during the creation of database and loaded befor the asset execution|

## Road Map

### Features

- [ ] Cross database references (comming soon)
- [ ] Custom Asset (Go)
- [ ] Tests
- [ ] Seeds
- [ ] Database Sources
- [ ] Custom SQL
- [ ] Pre/Post-hooks
- [ ] Embedded UI Dashboard

### Database support

- [x] DuckDB
- [ ] PostgreSQL (comming soon)
- [ ] MySQL
- [ ] ClickHouse
- [ ] SnowFlake

### Workflow

- [x] Channel
- [ ] Temporal.io
- [ ] Kafka Distibuted
