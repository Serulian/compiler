$module('loop', function () {
  var $static = this;
  $static.DoSomething = function (i) {
    return $t.fastbox(i.$wrapped + 1, $g.________testlib.basictypes.Integer);
  };
  $static.TEST = function () {
    var counter;
    var result;
    var stream;
    var $current = 0;
    syncloop: while (true) {
      switch ($current) {
        case 0:
          counter = $t.fastbox(0, $g.________testlib.basictypes.Integer);
          stream = $g.________testlib.basictypes.IntStream.OverRange($t.fastbox(1, $g.________testlib.basictypes.Integer), $t.fastbox(2, $g.________testlib.basictypes.Integer));
          $current = 1;
          continue syncloop;

        case 1:
          result = stream.Next();
          if (!$t.assertnotnull(result.Second).$wrapped) {
            $current = 2;
            continue syncloop;
          } else {
            $current = 4;
            continue syncloop;
          }
          break;

        case 2:
          $current = 3;
          continue syncloop;

        case 3:
          return $t.fastbox(counter.$wrapped == 5, $g.________testlib.basictypes.Boolean);

        case 4:
          counter = $t.fastbox(counter.$wrapped + $g.loop.DoSomething($t.syncnullcompare(result.First, function () {
            return $t.fastbox(0, $g.________testlib.basictypes.Integer);
          })).$wrapped, $g.________testlib.basictypes.Integer);
          $current = 1;
          continue syncloop;

        default:
          return;
      }
    }
  };
});
