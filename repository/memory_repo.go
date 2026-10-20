package repository
import (
	"sync"
)
type MemoryRepo struct {
	mu       sync.RWMutex
	users    map[string]string // Simplificado: username -> role
	vehicles map[int]string    // Simplificado: id -> status (ej. "Pendiente a Aprobado")
}
func NewMemoryRepo() *MemoryRepo {
	return &MemoryRepo{
		users:    make(map[string]string),
		vehicles: make(map[int]string),
	}
}
// Ejemplo de operación CRUD segura con Mutex
func (repo *MemoryRepo) AddVehicle(id int, status string) {
	repo.mu.Lock()
	defer repo.mu.Unlock()
	repo.vehicles[id] = status
}