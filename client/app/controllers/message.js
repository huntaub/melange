'use strict';

(function() {
  var melangeControllers = angular.module('melangeControllers');

  melangeControllers.controller('AllCtrl', ['$scope', 'mlgApi',
  function($scope, mlgApi) {
    // Sync up
    var sync = function() {
      $scope.loading = true;
      mlgApi.getMessages().$promise.then(function(data) {
        $scope.loading = false;
        $scope.newsfeed = data;
      });
    }
    sync();
    $scope.$on("mlgSyncApp", sync)

  }]);

  melangeControllers.controller('DashboardCtrl', ['$scope', 'mlgPlugins', 'mlgHelper', 'mlgTiles', 'mlgApi',
  function($scope, mlgPlugins, mlgHelper, mlgTiles, mlgApi) {
    // Tile Information
    $scope.editDash = false;
    mlgTiles.all().then(function(tiles) {
      $scope.tiles = tiles;
    })

    var stopFunc = function(e, ui) {
      console.log($scope.tiles);
    }

    // Conditionally enable tile building
    $scope.$watch(function() { return $scope.editDash }, function(data) {
      $scope.sortableOptions = {
        stop: stopFunc,
        disabled: !data,
      }
    })

    mlgPlugins.all().then(function(plugins) {
      $scope.plugins = plugins;
    });

    $scope.adding = false;
    $scope.addTile = function(plugin, tileKey) {
      $scope.tiles.push(mlgTiles.parse(plugin, tileKey));
      $scope.adding = false;
      mlgTiles.update($scope.tiles);
    }

    $scope.deleteTile = function(index) {
      $scope.tiles.splice(index, 1);
      mlgTiles.update($scope.tiles);
    }

    // Sync up if needed.
    var sync = function() {
      console.log("Syncing")
      $scope.loading = true;
      mlgApi.getMessages().$promise.then(function(data) {
        $scope.loading = false;
        $scope.newsfeed = data;
      });
    }
    sync();
    $scope.$on("mlgSyncApp", sync)

  }]);

  melangeControllers.controller('ProfileCtrl', ['$scope', 'mlgApi',
  function($scope, mlgApi) {
    $scope.newProfile = false;

    mlgApi.currentProfile().then(function(data) {
      console.log(data);
      $scope.profile = data;
    },
    function(err) {
      if(err === true) {
        $scope.newProfile = true;
      } else {
        console.log("Couldn't get profile. Something went wrong.")
        console.log(err)
      }
    });

    // Sync up if needed.
    var sync = function() {
      console.log("Syncing")
      $scope.loading = true;
      mlgApi.getMessages(true, false, false).$promise.then(function(data) {
        $scope.loading = false;
        $scope.newsfeed = data;
      });
    }
    sync();
    $scope.$on("mlgSyncApp", sync)

  }]);

  melangeControllers.controller('UserProfileCtrl', ['$scope', '$routeParams', 'mlgApi',
  function($scope, $routeParams, mlgApi) {
    $scope.profile = {
      name: $routeParams.alias
    }

    mlgApi.getMessage($routeParams.alias, "profile").then(function(profile) {
      console.log(profile)
      $scope.profile = {
        name: profile.components["airdispat.ch/profile/name"].string,
        description: profile.components["airdispat.ch/profile/description"].string,
        image: profile.components["airdispat.ch/profile/avatar"].string,
      }
    }, function(err) {
      console.log("Got an error getting profile")
      console.log(err)
    });

    // Sync up if needed.
    var sync = function() {
      console.log("Syncing")
      $scope.loading = true;
      mlgApi.getMessagesAtAlias($routeParams.alias).then(function(data) {
        $scope.loading = false;
        $scope.newsfeed = data;
      });
    }
    sync();
    $scope.$on("mlgSyncApp", sync)

  }]);

  melangeControllers.controller('NewProfileCtrl', ['$scope', '$location', 'mlgApi',
  function($scope, $location, mlgApi) {
    $scope.profile = {};
    $scope.save = function() {
      // Save the profile
      mlgApi.updateProfile($scope.profile).then(
        function() {
          $location.path("/profile");
        },
        function(err) {
          console.log("Error updating profile.")
          console.log(err)
        }
      )
    }
  }]);

})();
