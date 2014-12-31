/**
 * Copyright 2014 Google Inc. All rights reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

IOWA.Analytics = (function(exports) {

  "use strict";

  var GA_TRACKING_CODE = exports.DEV ? 'UA-58124138-2' : 'UA-58124138-1';

  /**
   * Analytics for Santa Tracker
   *
   * @constructor
   * @param {string} trackingCode GA tracking code.
   */
  function Analytics(trackingCode) {
    this.loadTrackingCode();

    ga('create', trackingCode, {
      'cookiePath': '/events/io2015',
      'siteSpeedSampleRate': 50 // 50% of users.
    });

    this.trackPageView(); // Track initial pageview.

    this.trackPerfEvent('template-bound', 'Polymer');
    this.trackPerfEvent('HTMLImportsLoaded', 'Polymer');
    this.trackPerfEvent('polymer-ready', 'Polymer');

    /**
     * A collection of timing categories, each a collection of start times.
     * @private {!Object<string, Object<string, ?number>}
     */
    this.startTimes_ = {};
  }

  Analytics.prototype.POLYMER_ANALYTICS_TIMEOUT_ = 60 * 1000;

  Analytics.prototype.loadTrackingCode = function() {
    (function(i,s,o,g,r,a,m){i['GoogleAnalyticsObject']=r;i[r]=i[r]||function(){
    (i[r].q=i[r].q||[]).push(arguments)},i[r].l=1*new Date();a=s.createElement(o),
    m=s.getElementsByTagName(o)[0];a.async=1;a.src=g;m.parentNode.insertBefore(a,m)
    })(window,document,'script','//www.google-analytics.com/analytics.js','ga');
  },

  /**
   * Tracks a page view. Page view tracking is throttled to prevent logging
   * page redirects by the URL router.
   * @param {string} opt_path Optional override page path to record.
   * @param {function} opt_callback Optional callback to be invoked after the
   *                   hit is recorded.
   */
  Analytics.prototype.trackPageView = function(opt_path, opt_callback) {
    var obj = {};
    if (opt_path) {
      obj.page = opt_path
    }
    if (typeof opt_callback === 'function') {
      obj.hitCallback = opt_callback
    }

    ga('send', 'pageview', obj);
  };

  /**
   * Tracks a performance timing. See
   * https://developers.google.com/analytics/devguides/collection/gajs/gaTrackingTiming#settingUp
   * @param {string} category Category of timing (e.g. 'Polymer')
   * @param {string} variable Name of the timing (e.g. 'polymer-ready')
   * @param {number} time Time, in milliseconds.
   * @param {string=} opt_label An optional sublabel, for e.g. A/B test identification.
   * @param {number=} opt_maxTime An optional max time, after which '- outliers' will be appended to variable name.
   * @param {object=} opt_obj Optional field object for additional params to send to GA.
   */
  Analytics.prototype.trackPerf = function(category, variable, time, opt_label, opt_maxTime, opt_obj) {
    if (opt_maxTime !== null && time > opt_maxTime) {
      variable += ' - outliers';
    }
    ga('send', 'timing', category, variable, time, opt_label, opt_obj);
  };

  /**
   * Tracks an event
   *
   * @param {string} category
   * @param {string} action
   * @param {string=} opt_label
   * @param {(string|number)=} opt_value
   */
  Analytics.prototype.trackEvent = function(category, action, opt_label, opt_value) {
    ga('send', 'event', category, action, opt_label, opt_value, {useBeacon: true});
  };

  /**
   * Tracks a social action
   *
   * @param {string} network
   * @param {string} action
   * @param {string} target
   */
  Analytics.prototype.trackSocial = function(network, action, target) {
    ga('send', 'social', network, action, target);
  };

  /**
   * Log Polymer startup performance numbers.
   */
  Analytics.prototype.trackPerfEvent = function(eventName, categoryName) {
    // performance.now() is sadly disabled even in some very recent browsers
    // TODO(bckenny): for now, only do polymer perf analytics in browsers with it.
    if (!(exports.performance && exports.performance.now)) {
      return;
    }

    document.addEventListener(eventName, function() {
      var now = exports.performance.now();

      if (exports.DEV) {
        console.info(eventName, '@', now);
      }

      this.trackPerf(categoryName, eventName, now, null,
                     this.POLYMER_ANALYTICS_TIMEOUT_, {'page': location.pathname});
    }.bind(this));
  };

  /**
   * Stores a start time associated with a category and variable name. When an
   * end time is registered with matching variables, the time difference is
   * sent to analytics. Use unique names if a race condition between timings is
   * possible; if a start time with the same names is registerd without an end
   * time in between, the original start time is discarded.
   * @param {string} category Category of timing (e.g. 'Assets load time')
   * @param {string} variable Name of the timing (e.g. 'polymer-ready')
   * @param {number} timeStart A timestamp associated with start, in ms.
   */
  Analytics.prototype.timeStart = function(category, variable, timeStart) {
    var categoryTimes = this.startTimes_[category] || (this.startTimes_[category] = {});
    categoryTimes[variable] = timeStart;
  };

  /**
   * Ends a timing event. The difference between the time associated with this
   * event and the timeStart event with the matching category and variable names
   * is sent to analytics. If no match can be found, the time is discarded.
   * @param {string} category Category of timing (e.g. 'Assets load time')
   * @param {string} variable Name of the timing (e.g. 'polymer-ready')
   * @param {number} timeEnd A timestamp associated with end, in ms.
   * @param {string=} opt_label An optional sublabel, for e.g. A/B test identification.
   * @param {number=} opt_maxTime An optional max time, after which '- outliers' will be appended to variable name.
   */
  Analytics.prototype.timeEnd = function(category, variable, timeEnd, opt_label, opt_maxTime) {
    var categoryTimes = this.startTimes_[category];
    if (!categoryTimes) {
      return;
    }
    var timeStart = categoryTimes[variable];
    if (timeStart !== null) {
      this.trackPerf(category, variable, timeEnd - timeStart, opt_label, opt_maxTime);
      categoryTimes[variable] = null;
    }
  };

  return new Analytics(GA_TRACKING_CODE);

})(window);
