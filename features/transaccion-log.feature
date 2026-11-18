# language: es
Característica: Integridad de transacción y log de datos (US-04)
  Como auditor del banco
  Quiero que cada operación quede registrada en un log JSON con su resultado
  Para garantizar la trazabilidad e integridad de las transacciones

  Regla: Resultado "exito" = operación aplicada; "fallo" = rechazada por regla de negocio; "error" = entrada ilegible o falla técnica

  Escenario: Retiro exitoso queda registrado como éxito
    Dado que Alice tiene 234.56 en su cuenta
    Cuando Alice intenta retirar 200.00
    Entonces el log registra una transacción "retiro" con resultado "exito"
    Y el registro indica saldo_antes 234.56 y saldo_despues 34.56

  Escenario: Retiro por encima del saldo queda registrado como fallo
    Dado que Hamza tiene 198.76 en su cuenta
    Cuando Hamza intenta retirar 200.00
    Entonces el log registra una transacción "retiro" con resultado "fallo"
    Y el saldo de Hamza sigue siendo 198.76

  Escenario: Abono exitoso queda registrado como éxito
    Dado que Alice tiene 234.56 en su cuenta
    Cuando Alice abona 100.00
    Entonces el log registra una transacción "abono" con resultado "exito"

  Escenario: Monto no numérico queda registrado como error
    Dado que Alice tiene 234.56 en su cuenta
    Cuando Alice ingresa el monto "abc" para retirar
    Entonces el log registra una transacción "retiro" con resultado "error"
    Y el saldo de Alice sigue siendo 234.56

  Escenario: Si el log no puede escribirse, la operación se revierte (atomicidad)
    Dado que Alice tiene 234.56 en su cuenta
    Y el log no puede escribirse
    Cuando Alice intenta retirar 200.00
    Entonces la operación termina con error
    Y el saldo de Alice sigue siendo 234.56

  Escenario: Cada línea del log es un JSON válido y los id son consecutivos
    Dado que Alice realizó un abono y dos retiros
    Entonces el log tiene 3 líneas JSON válidas con id 1, 2 y 3
