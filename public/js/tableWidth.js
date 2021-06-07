var tableAdjust = {
  ta:
    function tableAdjust(pagePct, topElements, topFixedExtra, headRowId, headTableId, scrollId, lastColId) {
      var topFixedHeight = topFixedExtra;

      topElements.forEach((val) => {
        oh = $(val).outerHeight();
        oht = $(val).outerHeight(true);
        topFixedHeight += oh + (oht - oh)/2;
      });

      $wh = $(window).innerHeight();
      $thh = $(headTableId).outerHeight(true);
      scrollerMaxHeight = pagePct * $wh - (topFixedHeight + $thh);
      if ($(scrollId).height() > scrollerMaxHeight) {
        $(scrollId).height(scrollerMaxHeight);
        adjustForScrollbar = 1;
      } else {
        adjustForScrollbar = 0;
      }

      $(headRowId).children('th').each(function() {
        $mw = 0;
        $className = $(this).attr('id');
        $('.' + $className).each(function() {
          $mw = Math.max($mw, $(this).width() + 1);
        });
        $('.' + $className).each(function() {
          $(this).width($mw);
        });
      });
      $tw = 0;
      $(headRowId).children('th').each(function() {
        $tw += $(this).outerWidth();
      });
      if ( adjustForScrollbar == 1 ) {
        $sw = getScrollBarWidth();
      } else {
        $sw = 0;
      }
      $(scrollId).width($tw+$sw);
      $(lastColId).width($(lastColId).width() + $sw);
      $(lastColId).attr('margin-right', $sw);
    }
}
