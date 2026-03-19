El programa lee los datos del fichero inquilinos.csv con el siguiente formato:

NOMBRE,DNI,TELÉFONO,INMUEBLE,RENTA,FECHA INICIO,RENTA INICIO,EMPRESA,COMENTARIOS

Se puede exportar la hoja con los datos de garajes de Google Sheet directamente en formato .csv con comas, copiarlo en la carpeta del programa y ponerle de nombre datos.csv

configuracion.json
{
    "arrendador": "Nombre del arrendador",
    "DNI": "DNI del arrendador",
    "mes": "2026-03",
    "id": 1
}

Compilado con:
go build -ldflags "-s -w"
