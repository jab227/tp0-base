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

La resolución de todos los ejercicios de la primera parte se encuentran en la branch `parte-1-docker`.

#### Ejercicio 1
En primer lugar presentamos un ejemplo de uso del script de generacion de DockerCompose:

```bash
$ ./generar-compose.sh docker-compose-dev.yaml 5
```
esto creara un docker compose con el nombre `docker-compose-dev.yaml`, con cinco clientes (`client1`, `client2`,...,`client5`)
La ejecución resulta en el siguiente archivo

*(AGREGAR-EJECUCION)*
```yaml

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

Se modifico tambien el script de generacion de DockerCompose para soportar volumenes.

#### Ejercicio 3

Se utilizara el script `validar-echo-server.sh` para interactuar con
el EchoServer con el comando `netcat`. Este script envia distintos
mensajes al servidor y prueba que la respuesta sea igual a lo que
envió. Si ocurre algún problema al conectarse con el mismo, o recibe
algo distinto a lo que envió, lo informara por _stdout_.

Para poder acceder a la red del servidor sin exponer puertos, se creo
un nuevo docker-compose, `docker-compose-netcat.yaml` y un nuevo
Dockerfile `Dockerfile_nc`. En el compose se crea un servicio `nctest`
que es el encargado de correr el script y depende del servicio del
server Un ejemplo de ejecución

*(AGREGAR EJECUCION)*
```bash

```

Por ultimo se agregaron nuevos targets al Makefile (siguiendo las
convenciones de los target originales) para facilitar la ejecucion de
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

En el protocolo diseñado todos los mensajes tienen esta forma

```
   1 Byte      Variable Length
+------------+-----------------+
|MESSAGE_KIND| MESSAGE_PAYLOAD |
+------------+-----------------+
```

done los valores posibles de *MESSAGE_KIND* son
```
+---------------------+-------+
| MESSAGE_KIND        | Value |
+---------------------+-------+
| BET                 |   0   |
+---------------------+-------+
| ACKNOWLEDGE         |   1   |
+---------------------+-------+
| DONE                |   2   |
+---------------------+-------+
| WINNERS             |   3   |
+---------------------+-------+
| WINNERS_UNAVAILABLE |   4   |
+---------------------+-------+
| WINNERS_LIST        |   5   |
+---------------------+-------+
```

#### Ejercicio 5
Para ver el código asociado a este ejercicio ver el branch `parte-2-ejercicio-5`

Para publicar una apuesta se manda el mensaje de *BET* al servidor

```
   1 Byte    4 Bytes        4 Bytes        PAYLOAD_SIZE bytes
+----------+--------------+--------------+---------------------+
| KIND=BET | PAYLOAD_SIZE |  AGENCY_ID   |  PAYLOAD            |
+----------+--------------+--------------+---------------------+
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
`parte-2-ejercicio-6`. Además es necesario descomprimir los contenidos
del archivo `.data/datasets.zip` en el root del proyecto, si se quiere
ejecutar los clientes.

Ahora se requiere leer las apuestas desde archivos provistos por la
cátedra, y que en una misma consulta se puedan enviar múltiples
apuestas. El tamaño de los batchs es configurable a través de la
variable de entorno:

- `CLI_BATCH_SIZE=${Size}`

si no la encuentra el valor por defecto es 16. Un punto a tener en
cuenta es que si el tamaño del mensaje a enviar superara los 8kB
(teniendo en cuenta los bytes asociados al tipo de mensaje) puede
ocurrir que se envié una cantidad menor de apuestas. Esto también
sucede si la cantidad de apuestas total no es divisible por la
cantidad de batchs.

Soportar el envió de múltiples apuestas en un request, requirió
modificar el mensaje de *BET*, el cual toma ahora la siguiente forma:

```
  1 Byte     4 Byte         4 Byte         PAYLOAD_SIZE Bytes
+----------+--------------+---------------+----------------------+
| KIND=BET | PAYLOAD_SIZE |   BET_COUNT   | PAYLOAD              |
+----------+--------------+---------------+----------------------+
```

Se agrega ahora el campo *BET_COUNT* que cuenta cuantas apuestas
fueron enviadas, nuevamente este campo es un *uint32* en _Little
Endian_. También cambia como se serializa el payload, ahora toda
apuesta va a estar delimitada por un `\n`, ignorando la ultima.


```
NOMBRE_1,APELLIDO_1,DOCUMENTO_1,NACIMIENTO_1,NUMERO_1\n
NOMBRE_2,APELLIDO_2,DOCUMENTO_2,NACIMIENTO_2,NUMERO_2\n
NOMBRE_3,APELLIDO_3,DOCUMENTO_3,NACIMIENTO_3,NUMERO_3\n
...
NOMBRE_N,APELLIDO_N,DOCUMENTO_N,NACIMIENTO_N,NUMERO_N\n
```

El mensaje de *ACKNOWLEDGE* tambien se modifico

```
     1 Byte          4 Bytes         BET_COUNT * 4 Bytes
+------------------+--------------+----------------------+
| KIND=ACKNOWLEDGE | BET_COUNT    | BET_NUMBERS          |
+------------------+--------------+----------------------+
```

Donde al igual que en *BET* se agrega el campo *BET_COUNT*, y también
el campo *BET_NUMBERS* el cual es una lista de uint32 codificados en
_Little Endian_ de longitud *BET_COUNT*.

Por ultimo se agrego el mensaje *DONE* para que el cliente señale que se envió el
ultimo batch.
```
    1 Byte
+-----------------+
| KIND=DONE       |
+-----------------+
```
Una vez recibido el servidor cierra la conexión.

Se agrega también la posibilidad de controlar el timeout de los
sockets del cliente a través de la variable de entorno.

- `CLI_SOCKET_TIMEOUT=15s `

La misma espera siempre un valor.

#### Ejercicio 7

Para ver el código asociado a este ejercicio ver el branch
`parte-2-ejercicio-7`

Ahora los clientes van a poder consultarle al servidor por la lista de
ganadores correspondiente a la agencia asociada, inmediatamente
después de enviar el mensaje de *DONE*. Para realizar esta consulta se
agrega un nuevo mensaje, *WINNERS*

```
   1 Byte            4 Bytes
+---------------+--------------------+
| KIND=WINNERS  |    AGENCY_ID       |
+---------------+--------------------+

```

En este mensaje se envía el ID de la agencia que esta solicitando los
ganadores como un *uint32* codificado en _Little Endian_. El mismo
campo se le agrega al mensaje de *DONE*

```
   1 Byte            4 Bytes
+---------------+--------------------+
| KIND=DONE     |    AGENCY_ID       |
+---------------+--------------------+
```


Antes de poder contestar este mensaje el servidor debe confirmar que
todas las agencias registraron todas sus apuestas. Cualquier consulta
por los ganadores previa a este escenario, recibirá como respuesta el
mensaje *WINNERS_UNAVAILABLE*.

```
    1 Byte
+--------------------------------+
| KIND=WINNERS_UNAVAILABLE       |
+--------------------------------+
```

Una vez que se tienen todas las apuestas de todas las agencias, el
servidor puede contestar a el mensaje de *WINNERS* con el siguiente
mensaje

```
      1 Byte            4 Bytes       4 Bytes        PAYLOAD_SIZE Bytes
+-------------------+---------------+--------------+--------------------------+
| KIND=WINNERS_LIST | WINNERS_COUNT | PAYLOAD_SIZE | PAYLOAD                  |
+-------------------+---------------+--------------+--------------------------+
```

done tanto *WINNERS_COUNT* como *PAYLOAD_SIZE* son uint32 codificados
en _Little Endian_.  El payload en este caso se corresponde a los DNIs
de los ganadores serializados de la siguiente forma

```
DNI_1,DNI_2,DNI_3,...,DNI_N
```

El cliente deberá reintentar la consulta si recibe un
*WINNERS_UNAVAILABLE*. Para esto se implementa una estrategia de
backoff configurable a través de las siguientes variables de entorno:

- `CLI_MAX_RETRIES=${number}`
- `CLI_BACKOFF=${duration}`

Una vez recibido todos los *DONE* desde las agencias, el servidor
utilizara las funciones `load_bets(...)` y `has_won(...)` provistas
por la cátedra para obtener y cachear los ganadores de cada agencia.

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
