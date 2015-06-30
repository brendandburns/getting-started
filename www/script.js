var app = angular.module('StarterApp', ['ngMaterial']);

app.controller('AppCtrl', ['$scope', '$mdSidenav', function($scope, $mdSidenav){
    $scope.data = {
	"selected": 0,
        "clusters": [
            {
		"name": "Create a new GKE Cluster"
            },
	    {
		"name": "Manual Cluster"
	    }
	]
    };
}]);