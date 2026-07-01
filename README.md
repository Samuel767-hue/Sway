# Sway

Aplicación de chat en tiempo real con Go y WebSocket.

## Despliegue en GitHub Pages

La interfaz estática del proyecto se publica automáticamente desde la rama `main` mediante GitHub Actions.

### Pasos para activar Pages
1. Ve a la configuración del repositorio en GitHub.
2. En la sección "Pages", selecciona "GitHub Actions" como fuente.
3. Haz push a la rama `main` o ejecuta el workflow manualmente.

### Desarrollo local
- Abre `index.html` en un navegador o sirve la carpeta con un servidor estático.
- El backend Go corre con:
  ```bash
  go run .
  ```
