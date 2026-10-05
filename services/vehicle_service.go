package services
import "automarket-backend/repository"

type VehicleService struct {
	repo *repository.MemoryRepo
}
func NewVehicleService(repo *repository.MemoryRepo) *VehicleService {
	return &VehicleService{repo: repo}
}
func (s *VehicleService) ApproveVehicle(id int) {
	// Lógica para pasar de "Pendiente" a "Aprobado"
	s.repo.AddVehicle(id, "Aprobado")
}