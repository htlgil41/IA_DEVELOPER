
# MCP: Model Context Protocol

## ¿Qué es MCP?

Respuesta rápida: **MCP (Model Context Protocol)** es un estándar abierto para conectar aplicaciones de IA con tus sistemas. La documentación oficial lo compara con un **puerto USB-C para aplicaciones de IA**: en vez de escribir una integración a medida por cada combinación de asistente y herramienta, cada lado habla el protocolo una vez.

Un servidor MCP ofrece tres cosas:

1. **Herramientas**: Funciones que el modelo ejecuta.
2. **Recursos**: Datos que el modelo lee.
3. **Prompts**: Plantillas predefinidas.

La comunicación se basa en **JSON-RPC 2.0**, a través de entrada/salida estándar (`stdio`) si el servidor corre en local, o vía HTTP si reside en la nube.

---

## 🏗️ El Problema y la Solución

* **El Problema:** Integrar 3 aplicaciones con 4 herramientas requiere 12 integraciones hechas a mano.
* **La Solución:** Con MCP, cada una habla el protocolo una vez, reduciendo la complejidad a solo 7 conexiones.

---

## 📚 Lo que aprenderás en este curso

### Fundamentos

* Los tres participantes: **Host, Client, Server** y por qué el host crea un cliente por cada servidor.
* Las dos capas: datos (JSON-RPC 2.0) y transporte.
* Las tres primitivas: Herramientas (ACCIÓN) vs. Recursos (LECTURA).
* Anatomía de una conversación: `initialize`, `capabilities`, `tools/list` y `tools/call`.

### Implementación y Go

* **¿Por qué Go?**: El SDK oficial es estable (v1), mantenido en colaboración con Google, y genera el esquema JSON de tus herramientas directamente desde tu `struct`. En otros lenguajes, este esquema se escribe a mano y se desincroniza constantemente.
* El servidor completo: compilado y probado con el **MCP Inspector**.
* Despliegue: de local a la nube cambiando una línea.

### Integración y Operación

* Conexión al asistente: en Go es solo la ruta del binario; en Python es un proceso más complejo (gestor, entorno, intérprete).
* **Tres servidores que se pagan solos**: tus logs, tu base de datos de solo lectura y tu API interna.

---

## ⚠️ Avisos Importantes

1. **Seguridad**: Un servidor MCP corre con tus permisos de sistema. No expongas todas tus funciones ciegamente.
2. **No es para todo**: No intentes forzar MCP en tareas que no lo requieren. Evalúa si la complejidad aporta valor.
3. **La Pregunta Honesta**: ¿Es esto solo otra forma de hacer llamadas a funciones (function calling)? Analizamos qué cambia de verdad y por qué es un salto cualitativo.

**Nota**: Podemos probar las funciones mediante npx @modelcontextprotocol/inspector ./BINARIO
