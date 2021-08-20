node {

	stage('Checkout') {

		checkout scm

	}

	stage('Archive') {

		def timestamp = new Date().format('yyyyMMdd')

		zip archive: true, zipFile: "$timestamp-output.zip"

	}
	stage('Deploy') {

		def remote = [:]
		remote.name = 'test'
		remote.host = 'ssh-server'
		remote.user = 'test'
		remote.password = 'test'
		remote.allowAnyHosts = true
		sshCommand remote: remote, command: "ls -lrt"
		//sshCommand remote: remote, command: "for i in {1..5}; do echo -n \"Loop \$i \"; date ; sleep 1; done"
		sshPut remote: remote, from: "$timestamp-output.zip", into: '.'
	
	}
	
}