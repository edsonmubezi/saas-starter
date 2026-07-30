package saga

import (
	"context"
	"fmt"
	"log"
)

// Step represents a single step in a saga
type Step struct {
	Name       string
	Execute    func(ctx context.Context) error
	Compensate func(ctx context.Context) error
}

// Saga coordinates distributed transactions using the Saga pattern
type Saga struct {
	steps []Step
}

// New creates a new saga
func New() *Saga {
	return &Saga{
		steps: []Step{},
	}
}

// AddStep adds a step to the saga
func (s *Saga) AddStep(name string, execute, compensate func(ctx context.Context) error) *Saga {
	s.steps = append(s.steps, Step{
		Name:       name,
		Execute:    execute,
		Compensate: compensate,
	})
	return s
}

// Execute runs all saga steps. If any step fails, it runs compensating transactions
// in reverse order to maintain consistency.
func (s *Saga) Execute(ctx context.Context) error {
	executedSteps := []Step{}

	// Execute forward
	for i, step := range s.steps {
		log.Printf("Saga: Executing step %d/%d: %s", i+1, len(s.steps), step.Name)

		if err := step.Execute(ctx); err != nil {
			log.Printf("Saga: Step %s failed: %v. Starting compensation...", step.Name, err)

			// Compensate in reverse order
			if compensateErr := s.compensate(ctx, executedSteps); compensateErr != nil {
				// Compensation failed - this is a critical error
				return fmt.Errorf("saga step %s failed: %w, and compensation also failed: %v",
					step.Name, err, compensateErr)
			}

			return fmt.Errorf("saga failed at step %s: %w (compensated successfully)", step.Name, err)
		}

		executedSteps = append(executedSteps, step)
	}

	log.Printf("Saga: All %d steps executed successfully", len(s.steps))
	return nil
}

// compensate runs compensating transactions in reverse order
func (s *Saga) compensate(ctx context.Context, executedSteps []Step) error {
	for i := len(executedSteps) - 1; i >= 0; i-- {
		step := executedSteps[i]

		if step.Compensate == nil {
			log.Printf("Saga: No compensation defined for step %s, skipping", step.Name)
			continue
		}

		log.Printf("Saga: Compensating step: %s", step.Name)

		if err := step.Compensate(ctx); err != nil {
			// Compensation failed - log but continue trying to compensate other steps
			log.Printf("Saga: Compensation failed for step %s: %v", step.Name, err)
			return fmt.Errorf("compensation failed for step %s: %w", step.Name, err)
		}
	}

	return nil
}

// Example usage:
// saga := saga.New()
// saga.AddStep("CreateEmployee",
//     func(ctx context.Context) error { return repo.CreateEmployee(...) },
//     func(ctx context.Context) error { return repo.DeleteEmployee(...) })
// saga.AddStep("CreateUser",
//     func(ctx context.Context) error { return userRepo.CreateUser(...) },
//     func(ctx context.Context) error { return userRepo.DeleteUser(...) })
//
// if err := saga.Execute(ctx); err != nil {
//     // Handle error - compensation already ran
// }