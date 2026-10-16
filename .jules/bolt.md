## 2025-05-10 - Grok mock member test timeouts with Python scripts
Learning: Mock CLI scripts written in Python (`fakeScript` using `python3`) have ~250ms process startup overhead in test environments. Tests configuring member timeouts below 250ms (such as `200*time.Millisecond`) can fail `Start()` during `initialize` or `session/new` RPC calls before the test prompt is even sent.
Action next time: Ensure test timeouts for mock process runners written in Python are set to at least `1*time.Second` to allow process startup while remaining fast enough for prompt hang timeout tests.
