var tableAdjust = {
  ta:
    function tableAdjust(pagePct, topElements, topFixedExtra, headRowId, headTableId, scrollId, lastColId) {
      var topFixedHeight = topFixedExtra;

      topElements.forEach((val) => {
	//alert(Array(val, $(val).outerHeight(true)));
	oh = $(val).outerHeight();
	oht = $(val).outerHeight(true);
	topFixedHeight += oh + (oht - oh)/2;
      });

      $wh = $(window).innerHeight();
      $thh = $(headTableId).outerHeight(true);
      $(scrollId).height(pagePct * $wh - (topFixedHeight + $thh));

      $(headRowId).children('th').each(function() {
	$mw = 0;
	$className = $(this).attr('id');
	$cells = $('.' + $className).each(function() {
	  $mw = Math.max($mw, $(this).width() + 1);
	});
	$cells = $('.' + $className).each(function() {
	  $(this).width($mw);
	});
      });
      $tw = 0
      $(headRowId).children('th').each(function() {
	$tw += $(this).outerWidth();
      });
      $sw = getScrollBarWidth();
      $(scrollId).width($tw+$sw);
      $(lastColId).width($(lastColId).width() + $sw);
      $(lastColId).attr('margin-right', $sw);
    }
}

$(document).ready(tableAdjust.ta(page_pct, top_elements, top_extra, head_row_id, head_table_id, scroller_id, last_col_id));
