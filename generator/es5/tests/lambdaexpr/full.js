$module('full', function () {
  var $static = this;
  $static.DoSomething = function () {
    var $state = {
      current: 0,
      returnValue: null,
    };
    $state.next = function ($callback) {
      try {
        while (true) {
          switch ($state.current) {
            case 0:
              function (firstParam, secondParam) {
                var $state = {
                  current: 0,
                  returnValue: null,
                };
                $state.next = function ($callback) {
                  try {
                    while (true) {
                      switch ($state.current) {
                        case 0:
                          1234;
                          $state.returnValue = 4567;
                          $state.current = -1;
                          $callback($state);
                          return;

                        default:
                          $state.current = -1;
                          return;
                      }
                    }
                  } catch (e) {
                    $state.error = e;
                    $state.current = -1;
                    $callback($state);
                  }
                };
                return $promise.build($state);
              };
              $state.current = -1;
              return;

            default:
              $state.current = -1;
              return;
          }
        }
      } catch (e) {
        $state.error = e;
        $state.current = -1;
        $callback($state);
      }
    };
    return $promise.build($state);
  };
});
