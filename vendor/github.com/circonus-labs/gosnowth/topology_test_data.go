package gosnowth

var topologyTestData = `[
	{"id":"1f846f26-0cfd-4df5-b4f1-e0930604e577","address":"10.8.20.1","port":8112,"apiport":8112,"weight":32,"n":2},
	{"id":"765ac4cc-1929-4642-9ef1-d194d08f9538","address":"10.8.20.2","port":8112,"apiport":8112,"weight":32,"n":2},
	{"id":"8c2fc7b8-c569-402d-a393-db433fb267aa","address":"10.8.20.3","port":8112,"apiport":8112,"weight":32,"n":2},
	{"id":"07fa2237-5744-4c28-a622-a99cfc1ac87e","address":"10.8.20.4","port":8112,"apiport":8112,"weight":32,"n":2}
]`

var topologyXMLTestData = `<nodes n="2">
	<node id="1f846f26-0cfd-4df5-b4f1-e0930604e577"
		address="10.8.20.1"
		port="8112"
		apiport="8112"
		weight="32"/>
	<node id="765ac4cc-1929-4642-9ef1-d194d08f9538"
		address="10.8.20.2"
		port="8112"
		apiport="8112"
		weight="32"/>
	<node id="8c2fc7b8-c569-402d-a393-db433fb267aa"
		address="10.8.20.3"
		port="8112"
		apiport="8112"
		weight="32"/>
	<node id="07fa2237-5744-4c28-a622-a99cfc1ac87e"
		address="10.8.20.4"
		port="8112"
		apiport="8112"
		weight="32"/>
</nodes>`
