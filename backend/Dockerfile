# Etapa de construcción
FROM golang:1.22-alpine AS build

WORKDIR /app

# Copia los archivos de go.mod y go.sum
COPY go.mod go.sum ./

# Descarga las dependencias necesarias
RUN go mod download

# Copia el resto de los archivos de la aplicación
COPY . .

# Construye la aplicación
RUN go build -o main .

# Etapa de producción
FROM alpine:latest

WORKDIR /app

# Copia el binario construido desde la etapa de construcción
COPY --from=build /app/main .

# Copia otros archivos necesarios
COPY config.json /app/config.json

# Expone el puerto
EXPOSE 1333

# Comando para ejecutar la aplicación
CMD ["./main"]
