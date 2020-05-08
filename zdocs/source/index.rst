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

	purchasePrice := 100
	previousPattern := patterns.DECREASING

	ticker := turnup.NewPriceTicker(purchasePrice, previousPattern)
	ticker.Prices[0] = 86

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

Now we can make some predictions based on our prices!

.. code-block:: go

	prediction, err := turnup.Predict(ticker)
    if err != nil {
        panic(err)
    }

	for _, potentialPattern := range prediction.Patterns {

		fmt.Println("Pattern:       ", potentialPattern.Pattern.String())
		fmt.Println("Progressions:  ", len(potentialPattern.PotentialWeeks))
		fmt.Printf("Chance:         %v%%\n", potentialPattern.Chance() * 100)
		fmt.Println("Min Guaranteed:", potentialPattern.MinPrice())
		fmt.Println("Max Potential: ", potentialPattern.MaxPrice())
		fmt.Println()

	}

Output:

.. code-block:: text

    Pattern:        BIG SPIKE
    Progressions:   7
    Chance:         85.59%
    Min Guaranteed: 200
    Max Potential:  600

    Pattern:        DECREASING
    Progressions:   1
    Chance:         9.51%
    Min Guaranteed: 85
    Max Potential:  90

    Pattern:        SMALL SPIKE
    Progressions:   7
    Chance:         4.9%
    Min Guaranteed: 140
    Max Potential:  200


.. note::

    If the ticker describes an impossible price pattern, it will be reported by ``err``
    and ``prediction`` will be ``nil``.

We can get some more information about specific potential price trends within each
over-arching pattern:

.. code-block:: go

	bigSpike, err := prediction.Pattern(patterns.BIGSPIKE)
	if err != nil {
		panic(err)
	}

	for _, potentialWeek := range bigSpike.PotentialWeeks {

		fmt.Printf("Chance: %v%%\n", potentialWeek.Chance() * 100)
		fmt.Println("Min Guaranteed:", potentialWeek.MinPrice())
		fmt.Println("Max Potential:", potentialWeek.MaxPrice())

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

    Chance: 12.23%
    Min Guaranteed: 200
    Max Potential: 600
    Monday AM: 85-90 (steady decrease)
    Monday PM: 90-140 (sharp increase)
    Tuesday AM: 140-200 (sharp increase)
    Tuesday PM: 200-600 (sharp increase)
    Wednesday AM: 140-200 (sharp decrease)
    Wednesday PM: 90-140 (sharp decrease)
    Thursday AM: 40-90 (random low)
    Thursday PM: 40-90 (random low)
    Friday AM: 40-90 (random low)
    Friday PM: 40-90 (random low)
    Saturday AM: 40-90 (random low)
    Saturday PM: 40-90 (random low)

Now get predicting!

Background Reading
==================

This library would not be possible without the `amazing work <https://gist.github.com/Treeki/85be14d297c80c8b3c0a76375743325b>`_
done by `Ninji <https://twitter.com/_Ninji>`_ and the
`in-depth breakdown <https://docs.google.com/document/d/1bSVNpOnH_dKxkAGr718-iqh8s8Z0qQ54L-0mD-lbrXo/edit>`_
of it by /u/Edricus. Both were intrumental in putting together this library and
/u/Edricus's breakdown is particular is recommended reading for any developers who want
to work on turnip price software.

.. _api documentation:

API documentation
=================

API documentation is created using godoc and can be
`found here <_static/godoc-root.html>`_.
