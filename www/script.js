var app = angular.module('StarterApp', ['ngMaterial'])
    .config(function($mdThemingProvider) {
	    $mdThemingProvider.theme('default')
	    .accentPalette('blue');
	});

var StarterApp = function(http) {
    this.http = http;
    this.clusters = {
        "create": {
		"name": "Create a new GKE Cluster"
            },
    };
    this.zones = [
	"asia-east1-a",
	"asia-east1-b",
	"asia-east1-c",
	"europe-west1-b",
	"europe-west1-c",
	"europe-west1-d",
	"us-central1-a",
	"us-central1-b",
        "us-central1-c",
	"us-central1-f"
    ];	
};

StarterApp.prototype.getPodJSON = function() {
    return {
	"apiVersion": "v1",
	"kind": "Pod",
	"metadata": {
	    "name": "nginx-pod",
	    "labels": {
		"role": "frontend"
	    }
	},
	"spec": {
	    "containers": [{
		"name": "nginx-container",
		"image": "nginx",
		"ports": [{
		    "containerPort": 80
		}]
	    }]
	}
    }
};

StarterApp.prototype.getReplicationController = function() {
    return {
	"apiVersion": "v1",
	"kind": "ReplicationController",
	"metadata": {
	    "name": "nginx-replication-controller",
	    "namespace": "default",
	},
	"spec": {
 	    "replicas": 3,
	    "selector": {
		"app": "nginx"
	    },
	    "template": {
		"metadata": {
		    "name": "nginx-pod",
		    "labels": {
			"app": "nginx"
		    },
		},
		"spec": {
		    "containers": [{
			"name": "nginx-container",
			"image": "nginx",
			"ports": [{
			    "containerPort": 80
			}]
		    }]
		}
	    }
	}
    }
};

StarterApp.prototype.getService = function() {
    return {
	"apiVersion": "v1",
	"kind": "Service",
	"metadata": {
	    "name": "nginx-service"
	},
	"spec": { 
	    "ports": [{
		"port": 80
	    }],
	    "selector": {
		"app": "nginx"
	    },
	    "type": "LoadBalancer"
	}
    }
};

StarterApp.prototype.LoadClusters = function() {
    if (!this.project || !this.zone) {
	return
    }
    return this.http.get('/list?project=' + this.project + '&zone=' + this.zone).
	success(function(data, status) {
	    angular.forEach(data.clusters, function(cluster) {
		this.clusters[cluster.name] = {"name": cluster.name};
	    }.bind(this));
	}.bind(this)).
        error(this.httpError);
};

StarterApp.prototype.SelectCluster = function(cluster) {
    return this.http.get('/select?project=' + this.project + '&zone=' + this.zone + '&cluster=' + cluster).
	success(function(data, status) {
	    this.selectedCluster = data;
	}.bind(this)).
        error(this.httpError);
};

StarterApp.prototype.httpError = function(data, status) {
    console.log("HTTP Error");
    console.log(data);
    console.log(status);
};

StarterApp.prototype.deploy = function() {
    var promise = this.createReplicationController();
    promise.then(function() {
        return this.createService();
    }.bind(this));
    return promise;
};

StarterApp.prototype.createCluster = function() {
    this.selectedCluster = null;
    return this.http.get("/create?project=" + this.project + "&zone=" + this.zone + "&cluster=" + this.clusterName).
        success(function(data, status) {
		this.deploying = true;
		this.selectedCluster = {
		    "name": this.clusterName,
		    "status": "PENDING",
		};
		this.clusters[this.clusterName] = this.selectedCluster;
		this.cluster = this.clusterName;
	    }.bind(this)).
        error(this.httpError);
};

StarterApp.prototype.createReplicationController = function() {
    return this.http.post("/api/v1/namespaces/default/replicationcontrollers", this.getReplicationController()).
	success(function(data, status) {
	    // TODO Toast here
	}.bind(this)).
        error(this.httpError);
};

StarterApp.prototype.createService = function() {
    return this.http.post("/api/v1/namespaces/default/services", this.getService()).
	success(function(data, status) {
	    // TODO Toast here
	}.bind(this)).
	error(this.httpError);
};

StarterApp.prototype.readyToDeploy = function() {
    return this.project && this.zone && this.cluster && !this.deployed() && this.selectedCluster && this.selectedCluster.status == "RUNNING";
};

StarterApp.prototype.readyToCreate = function() {
    return this.project && this.zone;
};

StarterApp.prototype.clusterDeploying = function() {
    return this.selectedCluster && this.selectedCluster.status != "RUNNING";
};

StarterApp.prototype.deployed = function() {
    return this.replicationController || this.service || (this.pods && this.pods.items.length > 0);
};

StarterApp.prototype.delete = function() {
    this.http.delete("/api/v1/namespaces/default/replicationcontrollers/" + this.replicationController.metadata.name).
        success(function(data, status) {
	    this.replicationController = null;
	}.bind(this)).
        error(this.httpError);
    this.http.delete("/api/v1/namespaces/default/services/" + this.service.metadata.name).
        success(function(data, status) {
	    this.service = null;
	}.bind(this)).
        error(this.httpError);
    angular.forEach(this.pods.items, function(pod) {
	    this.http.delete("/api/v1/namespaces/default/pods/" + pod.metadata.name).
		success(function(data, status) {
			this.replicationController = null;
		    }.bind(this)).
		error(this.httpError);
	}.bind(this));
    // TODO Combine the promises and error check.
    this.pods = null;
};

var makeLabelSelector = function(replicationController) {
    var result = [];
    angular.forEach(replicationController.spec.selector, function(value, key) {
	    result.push(key + "=" + value);
	})
    return result.join();
};

StarterApp.prototype.refresh = function() {
    var rc = this.getReplicationController();
    var svc = this.getService();
    if (this.readyToDeploy()) {
	this.http.get("/api/v1/namespaces/default/replicationcontrollers/" + rc.metadata.name).
	    success(function(data, status) {
	        this.replicationController = data;
	    }.bind(this)).
	    error(this.httpError);

	this.http.get("/api/v1/namespaces/default/services/" + svc.metadata.name).
	    success(function(data, status) {
	        this.service = data;
	    }.bind(this)).
	    error(this.httpError);

	this.http.get("/api/v1/namespaces/default/pods?labelSelector=" + makeLabelSelector(rc)).
	    success(function(data, status) {
	        this.pods = data;
	    }.bind(this)).
	    error(this.httpError);
    }
    if (this.selectedCluster) {
	this.SelectCluster(this.selectedCluster.name);
	this.deploying = this.selectedCluster.status != "RUNNING";
    }
};

app.controller('AppCtrl', ['$scope', '$mdSidenav', '$http', '$interval', function($scope, $mdSidenav, $http, $interval){
    $scope.data = {
	"selected": 0,
    };
    $scope.controller = new StarterApp($http);
    $interval($scope.controller.refresh.bind($scope.controller), 2500)
}]);