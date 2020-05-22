# About
This is a Discord bot that facilitates playing a small economy game where you can work for a fake currency, you can spend it to passively generate the currency, you can gamble it away and you can give it away to your friends or use it to try to steal their fake currency.

It currently has over 8 million transactions and 20,000 users. It has 1,000 active users and it runs on commodity hardware.

This project has been amazing in helping me learn poor Go architecture and I don't recommend you make the same mistakes I did here. Particularly, prioritize testing first. I should have wrapped all Slack and other 3rd party API interactions in an interface so I could wire up a test framework more easily. Had I done that from the get-go, it would have been trivial to continue to add features and test coverage as the project grew. Instead, myself and a friend had to do a massive rework to get the 3rd party interactions to a place where they didn't get in the way of core business logic. I recommend using an interface to wrap your 3rd party clients and then using a library like https://github.com/vektra/mockery to generate the mocks for an interface. Then you can use the mocks to inject expected behavior into the system in a test without depending on the 3rd party API interaction.

If you want to see more contemporary Go architecture please see this repo here: https://github.com/SophisticaSean/reddit-admin-bot
