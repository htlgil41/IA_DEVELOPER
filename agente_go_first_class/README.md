
# Construyendo Agentes de IA con Go

Un agente de IA es un bucle. Un bucle que llama a un modelo, mira si ha pedido usar una herramienta, la ejecuta y le devuelve el resultado.

Y hay una cosa que hay que entender antes que ninguna otra: **el modelo no ejecuta nada, solo lo pide; quien decide y ejecuta eres tú.**

En este curso montamos uno entero en Go con el kit de desarrollo oficial: declarar herramientas, el bucle con sus tres trampas, streaming, el prompt del sistema, contexto y coste, errores y límites, cómo se prueba, y cuándo NO deberías hacer un agente.

---

## 🚀 Lo que aprenderás

- **Fundamentos:** Qué es un agente y en qué se diferencia de una llamada normal.
- **Criterios de Diseño:** Los cuatro criterios oficiales antes de escribir uno (y el que casi nadie mira).
- **Implementación en Go:** La primera llamada al modelo desde Go, y por qué la respuesta es una lista de bloques.
- **Herramientas:** Declarar una herramienta: nombre, esquema, y la descripción que decide todo.
- **Arquitectura:** El bucle completo, con el identificador que tiene que cuadrar.
- **Gestión de Datos:** Por qué todos los resultados van en un solo mensaje.
- **Streaming:** El motivo técnico del streaming (más allá de la experiencia de usuario).
- **Prompt Engineering:** El prompt del sistema: las tres líneas que cambian el comportamiento.
- **Coste y Optimización:** Por qué el contexto se paga en cada vuelta, y cómo se arregla en tu función.
- **Observabilidad:** Qué registrar para poder depurar algo que no se reproduce.
- **Resiliencia:** Errores, reintentos y los dos límites que tienes que poner tú.
- **Ecosistema:** Herramientas propias o un servidor MCP: la regla para elegir.
- **Testing:** Cómo se testea un agente, y qué es una evaluación.

---

## ❓ Preguntas Frecuentes

### ¿Qué es exactamente un agente de IA?

Un bucle: llamas al modelo, miras si ha pedido usar una herramienta, la ejecutas y le devuelves el resultado, hasta que deja de pedir herramientas. El modelo nunca ejecuta nada por su cuenta.

### ¿Se puede hacer en Go, o hay que usar Python?

Se puede, y hay un kit de desarrollo oficial con tipos de Go. El ecosistema de IA está en Python, es cierto, pero lo que rodea al agente (un binario, concurrencia, tipos) es justo lo que Go hace bien. Y un agente es casi todo lo que lo rodea.

### ¿Cuándo NO debería hacer un agente?

Si la tarea se puede describir de antemano, si el resultado no justifica más coste y latencia, si el modelo no es capaz de ese tipo de tarea, o si no puedes detectar y deshacer un error. Con que falle uno de los cuatro, quédate en una llamada normal.

### ¿Por qué se dispara el coste de un agente?

Porque la API no tiene memoria: el historial lo mandas tú entero en cada vuelta. Una herramienta que devuelve mil líneas no cuesta una vez, cuesta en todas las vueltas que quedan.

### ¿Herramientas propias o un servidor MCP?

Si solo la usa tu agente, una función de Go. Si la va a usar algo que tú no controlas (tu editor, otro asistente), entonces un servidor MCP.

### ¿Un agente se puede testear si el modelo no es determinista?

Sí, partiéndolo en dos: las herramientas son funciones normales y se testean con una tabla; el bucle se prueba con un modelo de mentira que devuelve lo que tú decidas. Lo que no es un test es medir si el agente hace bien su trabajo con el modelo real: eso es una evaluación.
