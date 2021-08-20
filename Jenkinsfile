node {
	def timestamp
	
	stage('Checkout') {

		checkout scm

	}

	stage('Archive') {

		timestamp = new Date().format('yyyyMMdd-HH:mm:ss.SSS')

		zip archive: true, zipFile: "$timestamp-output.zip"

	}
	if (BRANCH_NAME == 'master') {
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
}