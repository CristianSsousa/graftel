package graftel

import "fmt"

// ErrInvalidConfig é retornado quando a configuração é inválida.
type ErrInvalidConfig struct {
	Field   string
	Message string
}

func (e *ErrInvalidConfig) Error() string {
	if e.Field != "" {
		return fmt.Sprintf("configuração inválida: campo '%s': %s", e.Field, e.Message)
	}
	return fmt.Sprintf("configuração inválida: %s", e.Message)
}

// ErrInitializationFailed é retornado quando a inicialização falha.
type ErrInitializationFailed struct {
	Component string
	Err       error
}

func (e *ErrInitializationFailed) Error() string {
	return fmt.Sprintf("falha ao inicializar %s: %v", e.Component, e.Err)
}

func (e *ErrInitializationFailed) Unwrap() error {
	return e.Err
}

// ErrShutdownFailed é retornado quando o shutdown falha.
type ErrShutdownFailed struct {
	Component string
	Err       error
}

func (e *ErrShutdownFailed) Error() string {
	return fmt.Sprintf("falha ao encerrar %s: %v", e.Component, e.Err)
}

func (e *ErrShutdownFailed) Unwrap() error {
	return e.Err
}
