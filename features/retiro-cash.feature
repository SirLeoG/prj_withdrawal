# language: es
Característica: Retiro de efectivo
  Como cliente titular de una cuenta
  Quiero retirar efectivo de mi cuenta
  Para disponer de mi dinero cuando lo necesite

  Regla: Los clientes no pueden retirar más de su saldo

  Escenario: Retiro exitoso dentro del saldo
    Dado que Alice tiene 234.56 en su cuenta
    Cuando Alice intenta retirar 200.00
    Entonces el retiro es exitoso
    Y el saldo de Alice es 34.56

  Escenario: Retiro rechazado por exceder el saldo
    Dado que Hamza tiene 198.76 en su cuenta
    Cuando Hamza intenta retirar 200.00
    Entonces el retiro es rechazado
    Y el saldo de Hamza sigue siendo 198.76

  Escenario: Retiro rechazado por monto inválido
    Dado que Alice tiene 234.56 en su cuenta
    Cuando Alice intenta retirar -50.00
    Entonces el retiro es rechazado
    Y el saldo de Alice sigue siendo 234.56
