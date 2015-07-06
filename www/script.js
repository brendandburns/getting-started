var app = angular.module('StarterApp', ['ngMaterial']);

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
	error(function(data, status) {
	    console.log("Error!");
	    console.log(status);
	});
};

StarterApp.prototype.SelectCluster = function(cluster) {
    return this.http.get('/select?project=' + this.project + '&zone=' + this.zone + '&cluster=' + cluster).
	success(function(data, status) {
	    console.log(data);
	}).
	error(function(data, status) {
	    console.log("Error!");
	    console.log(status);
	});
};

StarterApp.prototype.deploy = function() {
    this.http.post("/api/v1/namespaces/default/replicationcontrollers", this.getReplicationController()).
	success(function(data, status) {
	    console.log(data);
	}.bind(this)).
	error(function(data, status) {
	    console.log("Error!");
	    console.log(data);
	    console.log(status);
	});
    this.http.post("/api/v1/namespaces/default/services", this.getService()).
	success(function(data, status) {
	    console.log(data);
	}.bind(this)).
	error(function(data, status) {
	    console.log("Error!");
	    console.log(data);
	    console.log(status);
	});
};

StarterApp.prototype.readyToDeploy = function() {
    return this.project && this.zone && this.cluster;
};

app.controller('AppCtrl', ['$scope', '$mdSidenav', '$http', function($scope, $mdSidenav, $http){
    $scope.data = {
	"selected": 0,
    };
    $scope.controller = new StarterApp($http);
}]);