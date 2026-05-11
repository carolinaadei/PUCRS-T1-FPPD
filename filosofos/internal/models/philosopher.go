func (p *Philosopher) DineWithHierarchy(
	wg *sync.WaitGroup,
	iterations int,
) {
	defer wg.Done()

	for i := 0; i < iterations; i++ {

		p.Think()

		// Determine fork order
		firstFork := p.LeftFork
		secondFork := p.RightFork

		if p.RightFork.ID < p.LeftFork.ID {
			firstFork = p.RightFork
			secondFork = p.LeftFork
		}

		logger.Log(
			fmt.Sprintf(
				"Philosopher %d trying fork %d first",
				p.ID,
				firstFork.ID,
			),
		)

		<-firstFork.Token

		logger.Log(
			fmt.Sprintf(
				"Philosopher %d picked fork %d",
				p.ID,
				firstFork.ID,
			),
		)

		time.Sleep(100 * time.Millisecond)

		logger.Log(
			fmt.Sprintf(
				"Philosopher %d trying fork %d second",
				p.ID,
				secondFork.ID,
			),
		)

		<-secondFork.Token

		logger.Log(
			fmt.Sprintf(
				"Philosopher %d picked fork %d",
				p.ID,
				secondFork.ID,
			),
		)

		p.Eat()

		// Release forks
		firstFork.Token <- struct{}{}
		secondFork.Token <- struct{}{}

		logger.Log(
			fmt.Sprintf(
				"Philosopher %d released forks %d and %d",
				p.ID,
				firstFork.ID,
				secondFork.ID,
			),
		)
	}
}