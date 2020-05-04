Turnup-Go
==========

.. toctree::
   :maxdepth: 2
   :caption: Contents:

We at Nook, Nook and Nook are happy to announce Turnup-go! A golang library for
predicting turnip prices on your Animal Crossing island.

We would like to thank /u/Edricus and his  `fantastic breakdown <https://
docs.google.com/document/d/1bSVNpOnH_dKxkAGr718-iqh8s8Z0qQ54L-0mD-lbrXo
/edit#heading=h.cfdltvt5yfc2>`_ of `Ninji's work <https://gist.github.com/Treeki
/85be14d297c80c8b3c0a76375743325b>`_, both of which were instrumental in the creation of
this library.

Table of Contents
=================

* :ref:`quickstart`
* :ref:`api documentation`

.. _quickstart:

Quickstart
==========

To make a prediction, the first thing we need to do is set up a price ticker that will
store our island's prices for the current week. We'll say we bought our turnips for
100 bells on Sunday, and last week we had a fluctuating pattern:

.. code-block:: go

	// Make new price ticker with a sunday purchase price of 100 bells and previous
	// week's pattern of FLUCTUATING.
	purchasePrice := 100
	previousPattern := patterns.FLUCTUATING

	ticker := turnup.NewPriceTicker(purchasePrice, previousPattern)

Now we can add some price data. There are a few different ways we can set a price for a
given price period. All of the following operations add a price for the Monday Morning
Price:

.. code-block:: go

	// By price period index.
	//  0 = Monday Morning
	//  11 = Saturday Afternoon
	ticker.Prices[0] = 87

	// By weekday and time of day.
	ticker.SetPriceForDay(time.Monday, timeofday.AM, 87)

	// By time.Date.
	priceDate := time.Date(
		2020, 4, 6, 10, 0, 0, 0, time.UTC,
	)
	ticker.SetPriceForTime(priceDate, 87)

Now we can make some predictions based on our prices:

.. code-block:: go

	prediction := turnup.Predict(ticker)

	for _, potentialPattern := range prediction.Patterns {

		fmt.Println("Pattern:       ", potentialPattern.Pattern.String())
		fmt.Println("Progressions:  ", len(potentialPattern.PotentialWeeks))
		fmt.Printf("Chance:         %v%%\n", potentialPattern.Analysis().Chance * 100)
		fmt.Println("Min Guaranteed:", potentialPattern.Analysis().MinPrice())
		fmt.Println("Max Potential: ", potentialPattern.Analysis().MaxPrice())
		fmt.Println()

	}

Output:

.. code-block:: text

    Pattern:        BIG SPIKE
    Progressions:   7
    Chance:         61.72%
    Min Guaranteed: 200
    Max Potential:  600

    Pattern:        DECREASING
    Progressions:   1
    Chance:         30.86%
    Min Guaranteed: 85
    Max Potential:  90

    Pattern:        SMALL SPIKE
    Progressions:   7
    Chance:         7.41%
    Min Guaranteed: 140
    Max Potential:  200

    Pattern:        FLUCTUATING
    Progressions:   0
    Chance:         0%
    Min Guaranteed: 0
    Max Potential:  0

We can get some more information about specific potential price trends within each
over-arching pattern:

.. code-block:: go

	// Returns err if pattern is invalid
	bigSpike, err := prediction.Pattern(patterns.BIGSPIKE)
	if err != nil {
		panic(err)
	}

	for _, potentialWeek := range bigSpike.PotentialWeeks {

		fmt.Printf("Chance: %v%%\n", potentialWeek.Analysis().Chance * 100)
		fmt.Println("Min Guaranteed:", potentialWeek.Analysis().MinPrice())
		fmt.Println("Max Potential:", potentialWeek.Analysis().MaxPrice())

		for _, potentialPeriod := range potentialWeek.PricePeriods {

            fmt.Printf(
				"%v %v: %v-%v (%v)\n",
				potentialPeriod.PricePeriod.Weekday(),
				potentialPeriod.PricePeriod.ToD(),
				potentialPeriod.MinPrice(),
				potentialPeriod.MaxPrice(),
				potentialPeriod.PatternPhase.Name(),
			)

		}

		fmt.Println()
	}

Each potential price pattern for the week will give an output block like so:

.. code-block:: text

    Chance: 8.82%
    Min Guaranteed: 200
    Max Potential: 600
    Monday AM: 85-90 (steady decrease)
    Monday PM: 80-87 (steady decrease)
    Tuesday AM: 75-84 (steady decrease)
    Tuesday PM: 70-81 (steady decrease)
    Wednesday AM: 65-78 (steady decrease)
    Wednesday PM: 60-75 (steady decrease)
    Thursday AM: 55-72 (steady decrease)
    Thursday PM: 90-140 (sharp increase)
    Friday AM: 140-200 (sharp increase)
    Friday PM: 200-600 (sharp increase)
    Saturday AM: 140-200 (sharp decrease)
    Saturday PM: 90-140 (sharp decrease)

Now get predicting!

.. _api documentation:

API documentation
=================

API documentation is created using godoc and can be
`found here <_static/godoc-root.html>`_.
