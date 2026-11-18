# language: es
Característica: Abono en efectivo (US-01b)
  Como cliente titular de una cuenta
  Quiero abonar efectivo a mi cuenta
  Para incrementar mi saldo disponible

  Regla: Solo se aceptan abonos con monto mayor a cero

  Escenario: Abono exitoso
    Dado que Alice tiene 234.56 en su cuenta
    Cuando Alice abona 100.00
    Entonces el abono es exitoso
    Y el saldo de Alice es 334.56

  Escenario: Abono rechazado por monto cero
    Dado que Alice tiene 234.56 en su cuenta
    Cuando Alice abona 0.00
    Entonces el abono es rechazado
    Y el saldo de Alice sigue siendo 234.56

  Escenario: Abono rechazado por monto negativo
    Dado que Hamza tiene 198.76 en su cuenta
    Cuando Hamza abona -20.00
    Entonces el abono es rechazado
    Y el saldo de Hamza sigue siendo 198.76

  Escenario: Abono seguido de retiro (integración con Retiro)
    Dado que Hamza tiene 198.76 en su cuenta
    Cuando Hamza abona 50.00
    Y Hamza intenta retirar 200.00
    Entonces el retiro es exitoso
    Y el saldo de Hamza es 48.76
