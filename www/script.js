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

StarterApp.prototype.LoadClusters = function() {
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

app.controller('AppCtrl', ['$scope', '$mdSidenav', '$http', function($scope, $mdSidenav, $http){
    $scope.data = {
	"selected": 0,
    };
    $scope.controller = new StarterApp($http);
}]);