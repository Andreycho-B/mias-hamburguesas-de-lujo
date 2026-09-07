# Reglas Normativas de Redacción y Comunicación

Basado en `.standard/03. REGLAS DE REDACCIÓN.md` (Nivel 2) y `.standard/12. INGENIERÍA DE PROMPTS Y COMUNICACIÓN CON EL USUARIO.md`.

## 1. Densidad y Supresión de Relleno
- Eliminar frases vacías (*"Es importante mencionar"*, *"Cabe destacar"*, *"Vale la pena señalar"*).
- Eliminar transiciones artificiales (*"Dicho esto"*, *"En este contexto"*, *"Una vez aclarado lo anterior"*).
- Prohibidas las conclusiones forzadas (*"En resumen"*, *"En conclusión"*) que no aporten síntesis nueva.
- Prohibidos párrafos espejo (dos párrafos consecutivos reformulando la misma idea).

## 2. Supresión de Validación, Marketing y Tono Motivacional
- No validar emocionalmente al usuario (*"Entiendo perfectamente"*, *"Excelente pregunta"*, *"Tienes toda la razón"*).
- Cero lenguaje publicitario, superlativos no comprobables o entusiasmo simulado. El texto describe objetivamente; no vende.
- Cero lenguaje antropomórfico sobre el modelo (*"sé que"*, *"pienso que"*, *"siento que"*). Expresar la operación técnica con precisión.

## 3. Ruptura de Firmas Estadísticas de LLM
- Prohibida la estructura "No X, sino Y" repetida como plantilla argumentativa.
- Evitar simetría forzada en párrafos o listas cuando el contenido no lo requiera.
- Prohibidas introducciones genéricas de saludo (*"¡Claro!"*, *"¡Por supuesto!"*, *"Veamos"*).
- No estructurar toda respuesta como lista con viñetas si la narrativa o tabla técnica es más clara.

## 4. Rigor en Adjetivos y Absolutos
- Todo adjetivo técnico (*eficiente*, *robusto*, *escalable*) debe acompañarse de la propiedad observable que lo justifica.
- Evitar absolutos (*siempre*, *nunca*, *perfecto*, *imposible*) a menos que exista respaldo formal.

## 5. Separación Explícita de Categorías
- Diferenciar hechos comprobados de interpretaciones o propuestas utilizando marcadores explícitos:
  - `Dato:` Información objetiva verificada en el código o repositorio.
  - `Interpretación:` Deducción técnica basada en la evidencia.
  - `Recomendación:` Propuesta operativa o arquitectónica para decisión del usuario.
