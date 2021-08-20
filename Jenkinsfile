node {

	stage('Checkout') {

		checkout scm

	}

	stage('Archive') {

		def timestamp = new Date().format('yyyyMMdd')

		zip archive: true, zipFile: "$timestamp-output.zip"

	}

}