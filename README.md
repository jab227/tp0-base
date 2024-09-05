# TP0: Docker + Comunicaciones + Concurrencia
## Informe

Para ejecutar los distintos ejercicios, salvo que se aclare lo
contrario en la sección correspondiente, se ejecutan de la siguiente
	manera:
- `make docker-compose-up` para inicializar tanto clientes como
  servidor
- `make docker-compose-logs` para ver los logs de clientes y servidor
- `make docker-compose-down` para detener los containers

### Parte 1

#### Ejercicio 1
En primer lugar presentamos un ejemplo de uso del script de generación de DockerCompose:

```bash
$ ./generar-compose.sh docker-compose-dev.yaml 5
```
esto creara un docker compose con el nombre `docker-compose-dev.yaml`, con cinco clientes (`client1`, `client2`,...,`client5`)
La ejecución resulta en el siguiente archivo

```yaml
name: tp0
services:
  server:
    container_name: server
    image: server:latest
    entrypoint: python3 ./main.py
    environment:
      - PYTHONUNBUFFERED=1
      - LOGGING_LEVEL=DEBUG
    volumes:
      - ./server:/echoserver
    networks:
      - testing_net
  client1:
    container_name: client1
    image: client:latest
    entrypoint: /client
    environment:
      - CLI_ID=1
      - CLI_LOG_LEVEL=DEBUG
    networks:
      - testing_net
    depends_on:
      - server
    volumes:
      - ./client:/echoclient
  client2:
    container_name: client2
    image: client:latest
    entrypoint: /client
    environment:
      - CLI_ID=2
      - CLI_LOG_LEVEL=DEBUG
    networks:
      - testing_net
    depends_on:
      - server
    volumes:
      - ./client:/echoclient
  client3:
    container_name: client3
    image: client:latest
    entrypoint: /client
    environment:
      - CLI_ID=3
      - CLI_LOG_LEVEL=DEBUG
    networks:
      - testing_net
    depends_on:
      - server
    volumes:
      - ./client:/echoclient
networks:
  testing_net:
    ipam:
      driver: default
      config:
        - subnet: 172.25.125.0/24
```
#### Ejercicio 2

Para lograr que realizar cambios en los archivos de configuración no
requiera un rebuild del container se utilizo el mecanismo provisto por
docker, *volumes*, el cual permite persistir los archivos generados y
usados por los containers en la maquina host. En primer lugar se tuvo
que agregar un `WORKDIR` distinto de `/` tanto al Dockerfile del
cliente como del servidor, ya que un volumen no puede estar montado
sobre `/`. Luego se agrego a los servicios de los clientes y del
servidor la siguiente llave:

```yaml
volumes:
  - ./{client/server}:/{echoclient/echoserver}
```

Para ver que los cambios persisten basta con levantar el compose,
editar el config del cliente y/o servidor y reiniciar los servicios
con:

`docker compose -f docker-compose-dev.yaml restart`

Se modifico tambien el script de generación de DockerCompose para soportar volúmenes.

#### Ejercicio 3

Se utilizara el script `validar-echo-server.sh` para interactuar con
el EchoServer con el comando `netcat`. Este script envía distintos
mensajes al servidor y prueba que la respuesta sea igual a lo que
envió. Si ocurre algún problema al conectarse con el mismo, o recibe
algo distinto a lo que envió, lo informara por _stdout_.

Para poder acceder a la red del servidor sin exponer puertos, se creo
un nuevo docker-compose, `docker-compose-netcat.yaml` y un nuevo
Dockerfile `Dockerfile_nc`. En el compose se crea un servicio `nctest`
que es el encargado de correr el script y depende del servicio del
server.

Por ultimo se agregaron nuevos targets al Makefile (siguiendo las
convenciones de los target originales) para facilitar la ejecución de
esta prueba
	- `make docker-compose-up-nc` para levantar el contenedor
	- `make docker-compose-down-nc` para detener el contenedor
	- `make docker-compose-logs-nc` para ver los logs
	
#### Ejercicio 4

Para probar que tanto cliente como servidor terminan de forma
_graceful_ al recibir *SIGTERM*, primero iniciamos los containers y
abrimos los logs con los comandos vistos anteriormente. Luego desde
otra terminal ejecutamos:

`docker compose -f docker-compose-dev.yaml down -t 5 <nombre-servicio>`

para enviar el *SIGTERM* al servicio en cuestión (si no se especifica
se lo manda a todos), y deberíamos ver esto reflejado en los logs.

### Parte 2
En los siguientes ejercicios se modifico tanto el cliente como el
servidor que el caso de uso denominado *Loteria Nacional*. El
protocolo planteado fue evolucionando a medida que se agregaban
requerimientos en los distintos ejercicios.

En el protocolo diseñado todos los mensajes enviados desde el cliente
tienen esta forma

```
+-------+--------------------+----------------------+
|       |                    |                      |
| KIND  |     AGENCYID       |     PAYLOAD_SIZE     |
|       |                    |                      |
+-------+--------------------+----------------------+
|                                                   |
|               VARIABLE_PAYLOAD                    |
|                                                   |
+---------------------------------------------------+
```

y los enviados desde el servidor

```
+------------------------+--------------------------+
|                        |                          |
|          KIND          |       PAYLOAD_SIZE       |
|                        |                          |
+------------------------+--------------------------+
|                                                   |
|               VARIABLE_PAYLOAD                    |
|                                                   |
+---------------------------------------------------+
```

Los tamaños (en bytes) de los campos fijos son
	- *KIND*: 1 
	- *AGENCYID*: 4
	- *PAYLOAD_SIZE*: 4

y los mismos estan codificados en _little endian_.

Los valores posibles de *KIND* son para los requests del client
```
+---------------------+-------+
| KIND                | Value |
+---------------------+-------+
| POST_BET            |   0   |
+---------------------+-------+
| BET_BATCH           |   1   |
+---------------------+-------+
| BET_BATCH_END       |   2   |
+---------------------+-------+
| GET_WINNERS         |   3   |
+---------------------+-------+
```

mientras que los responses del servidor

```
+---------------------+-------+
| KIND                | Value |
+---------------------+-------+
| ACKNOWLEDGE         |   0   |
+---------------------+-------+
| WINNERS_READY       |   1   |
+---------------------+-------+
| BETTING_RESULTS     |   2   |
+---------------------+-------+
```

#### Ejercicio 5
Para ver el código asociado a este ejercicio ver el branch `ej5`

Para publicar una apuesta se manda el mensaje de *BET* al servidor

```
   1 Byte    4 Bytes        4 Bytes        PAYLOAD_SIZE bytes
+---------------+----------+--------------+-------------------+
| KIND=POST_BET | AGENCYID | PAYLOAD_SIZE |      PAYLOAD      |
+--------------+-----------+--------------+-------------------+
```

donde *PAYLOAD_SIZE* y *AGENCY_ID* son un *uint32* en _Little Endian__
que contiene el numero de apuesta.

Dentro del payload se encuentran los campos requeridos como strings
separados por comas:

```
NOMBRE,APELLIDO,DOCUMENTO,NACIMIENTO,NUMERO
```

Una vez procesada la apuesta el servidor responde con un mensaje
*ACKNOWLEDGE*

```
     1 Byte              4 bytes
 +------------------+--------------------+
 | KIND=ACKNOWLEDGE |   BET_NUMBER       |
 +------------------+--------------------+
```
donde  *BET_NUMBER* es un *uint32* que contiene el numero de apuesta en _Little Endian_.

Recibido el *ACKNOWLEDGE* el cliente procede a loggear que la
respuesta fue recibida.  En caso de que la respuesta no sea recibida
antes del timeout, el cliente considerara que la apuesta no pudo ser
procesada.

Si se quiere modificar los datos de las apuestas para alguno de los
clientes, se pueden modificar las siguientes variables de entornos en
el docker-compose:
- `CLI_BETTOR_NOMBRE=${Nombre}`
- `CLI_BETTOR_APELLIDO=${Apellido}`
- `CLI_BETTOR_DOCUMENTO=${Documento}`
- `CLI_BETTOR_NACIMIENTO=${Fecha de Nacimiento}`
- `CLI_BETTOR_NUMERO=${Numero de apuesta}`

#### Ejercicio 6
Para ver el código asociado a este ejercicio ver el branch
`ej6`. Además es necesario descomprimir los contenidos
del archivo `.data/datasets.zip` en la misma carpeta (`.data/datasets`), si se quiere
ejecutar los clientes.

Ahora se requiere leer las apuestas desde archivos provistos por la
cátedra, y que en una misma consulta se puedan enviar múltiples
apuestas. El tamaño de los batchs es configurable a través del campo
`maxAmount` en el archivo `config.yaml`

Un punto a tener en cuenta es que si el tamaño del mensaje a enviar
superara los 8kB (teniendo en cuenta los bytes asociados al tipo de
mensaje) puede ocurrir que se envié una cantidad menor de
apuestas. Esto también sucede si la cantidad de apuestas total no es
divisible por la cantidad de batchs.

Soportar el envió de múltiples apuestas en un request, requirió crear
el mensaje de *BET_BATCH*. Para serializar las distintas apuestas
dentro de un batch/chunk, se encapsularon las mismas dentro de un
paquete con la siguiente forma

```
 4 Byte 
+------+-------------+------+-------------+--------+------+-------------+
| SIZE | BET_PAYLOAD | SIZE | BET_PAYLOAD |  ....  | SIZE | BET_PAYLOAD |
+------+-------------+------+-------------+--------+------+-------------+
```

Este nuevo formato permite identificar donde comienza y termina una nueva apuesta.

Por ultimo se agrego el mensaje *BET_BATCH_END* para que el cliente señale que se envió el
ultimo batch.

Una vez recibido el servidor cierra la conexión.

Se agrega también la posibilidad de controlar el timeout de los
sockets del cliente a través de la variable de entorno.

- `CLI_SOCKET_TIMEOUT=15s `

La misma espera siempre un valor.

#### Ejercicio 7

Para ver el código asociado a este ejercicio ver el branch
`ej7`

Ahora los clientes van a poder consultarle al servidor por la lista de
ganadores correspondiente a la agencia asociada, inmediatamente
después de enviar el mensaje de *BET_BATCH_END*. Para realizar esta
consulta se agrega un nuevo mensaje, *GET_WINNERS*. 

Antes de poder contestar este mensaje el servidor debe confirmar que
todas las agencias registraron todas sus apuestas. Por lo tanto los
clientes deben esperar a que el servidor envie el mensaje de
*WINNERS_READY* antes de poder preguntar por los mismos. Para que no
esperen para siempre se implemento una estrategia de backoff.

El payload para *GET_WINNERS* se corresponde a los DNIs
de los ganadores serializados de la siguiente forma

```
DNI_1,DNI_2,DNI_3,...,DNI_N
```

### Parte 3
Para ver el código asociado a este ejercicio ver el branch `ej8`

En esta parte se pide que el servidor acepte y procese mensajes en
paralelo. Utilizando la librería *multiprocessing* de python se
implemento la siguiente solución. En primer lugar se crea un pool de
procesos al cual se enviara un nuevo trabajo por cada nueva conexión.
Este trabajo consistirá en ejecutar el protocolo. Debe existir un
estado compartido entre los distintos procesos, para llevar cuenta de
cuantas de las agencias terminaron de enviar sus apuestas, también se
debe controlar el acceso a la función `load_bets(...)` ya que la misma
no es thread-safe/process-safe. Se decidió por encapsular este estado
y este comportamiento en la clase `BetsStorage` la cual se encarga de
garantizar la serializabilidad de los accesos de lectura y escritura a
el recurso. La implementación original, consistía en garantizar la
serializabilidad utilizando un lock, `multiprocessing.Lock()`, pero
debido a limitaciones de la librería esto no esta permitido al
utilizar un pool. Por lo que se recurrió a crear un manager,
`multiprocessing.Manager`, el cual crea un procesos que se encarga de
manejar una versión centralizada de un objeto y proveer acceso al
mismo.


Ademas se utilizo una barrera `multiprocessin.Barrier()` de manera tal
de esperar a que todos los clientes hayan confirmado el envio de todas
las apuestas y se pueda comenzar la eleccion de ganadores.

# Notas
Algunas aclaraciones de cosas que se podrian mejorar:
	- Ciertos valores podrian ser variables de entorno para hacer al
      trabajo mas configurable
	- Mas tests para el servidor especificamente
