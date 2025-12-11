package orchestrator

import (
	"context"
	"math/rand"
	"time"

	"ai-service-go/internals/types"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

// WittyMessages contains all the interactive messages used during processing
type WittyMessages struct {
	Thinking      []string
	MultiStep     []string
	FinalAnswer   []string
	ErrorRecovery []string
	ToolStart     map[string][]string
	ToolEnd       map[string][]string
	LongWait      []string
	Greeting      []string
	TipsAndFacts  []string
	AnimatedDots  []string
	Encouragement []string
}

// DefaultWittyMessages returns the default set of witty messages
func DefaultWittyMessages() *WittyMessages {
	return &WittyMessages{
		Thinking: []string{
			// Curious & Engaged
			"🤔 Hmm, let me think about this...",
			"🧠 Processing your request... my circuits are warming up!",
			"💭 Interesting question! Let me ponder this for a moment...",
			"🔮 Consulting my crystal ball... just kidding, using actual data!",
			"⚡ Neurons firing... well, digital ones anyway!",
			"🎯 Zeroing in on your answer...",
			"📚 Flipping through my knowledge pages...",
			"🔍 Detective mode: ON",
			// Playful
			"🎩 Let me pull some answers out of my hat...",
			"🧪 Running this through my analysis lab...",
			"🎬 And... action! Processing your request...",
			"🚀 Launching into research mode...",
			"🌟 Channeling my inner genius...",
			"🔧 Cranking up the knowledge engine...",
			"🎪 Step right up! Let me work some magic...",
			"🧩 Time to solve this puzzle...",
			// Professional but fun
			"📋 On it! Let me gather the intel...",
			"💼 Putting on my thinking cap...",
			"🎯 Challenge accepted! Working on it...",
			"🌐 Scanning my databases...",
			"⚙️ Gears are turning...",
			"🔬 Analyzing your request...",
			"📡 Tuning into the right frequency...",
			"🗃️ Digging through the archives...",
		},

		MultiStep: []string{
			// Progress updates
			"🔄 Still working on it... good things take time!",
			"⏳ Gathering more info... almost there!",
			"🧩 Piecing together the puzzle...",
			"🎪 Juggling multiple sources of info here...",
			"🔗 Connecting the dots...",
			"📊 Crunching some more data...",
			"🎯 Getting closer to the answer...",
			"🧵 Following the thread...",
			// Encouraging
			"💪 Making progress! Hang tight...",
			"🏃 Running to fetch more details...",
			"🔎 Diving deeper...",
			"🌊 Deep diving for more info...",
			"🎨 Adding the finishing touches...",
			"🔨 Building your answer piece by piece...",
			"🧭 Navigating through the data...",
			"⛏️ Mining for more insights...",
			// Fun
			"🐕 Fetching more data... woof!",
			"🍳 Cooking up something good...",
			"🎵 Humming along while I work...",
			"🏗️ Construction in progress...",
		},

		FinalAnswer: []string{
			// Celebration
			"🎉 Got it! Here's what I found...",
			"✨ Ta-da! Your answer is ready!",
			"🎯 Bullseye! Found exactly what you need!",
			"✅ Mission accomplished! Here you go...",
			"💡 Eureka! Here's your answer!",
			"🏆 Nailed it! Here's the scoop...",
			"🌟 Voilà! Your answer awaits...",
			"🎁 Special delivery! One answer coming right up...",
			// Confident
			"👍 All done! Here's what I've got...",
			"🚀 Mission complete! Launching your answer...",
			"📬 Fresh from the knowledge oven!",
			"🎊 Success! Here's the lowdown...",
			"💎 Found the gem you were looking for!",
			"🔓 Unlocked the answer!",
			"📍 X marks the spot! Found it!",
			"🎪 And for my final trick... your answer!",
		},

		ErrorRecovery: []string{
			"🔧 Oops, hit a small snag. Let me try another approach...",
			"🤔 That didn't work as expected. Plan B incoming!",
			"💪 Minor setback, but I'm not giving up!",
			"🔄 Let me recalibrate and try again...",
			"🎯 Adjusting my aim... one more try!",
			"🧭 Slight detour, but I know another way!",
			"🔍 Hmm, let me look at this from a different angle...",
			"🛠️ Quick pit stop, then back on track!",
			"🌊 Rolling with the punches here...",
			"🎲 That path didn't work, trying another route!",
		},

		LongWait: []string{
			"☕ This is taking a moment... grab a coffee?",
			"🎵 *elevator music plays*",
			"⏰ Patience, young grasshopper...",
			"🐢 Slow and steady wins the race...",
			"🧘 Taking a mindful moment to process...",
			"📖 Rome wasn't built in a day, and neither is this answer...",
			"🎭 Behind the scenes magic happening...",
			"🌙 Good things come to those who wait...",
			"🎬 Loading... like a blockbuster movie!",
			"🧪 Complex question requires complex processing...",
		},

		Greeting: []string{
			"👋 Hey there! What can I help you with?",
			"🌟 Hello! Ready to assist!",
			"😊 Hi! How can I make your day easier?",
			"🎯 At your service! What do you need?",
			"💡 Hello! Let's solve some problems together!",
		},

		AnimatedDots: []string{
			"⏳ Working.",
			"⏳ Working..",
			"⏳ Working...",
			"⌛ Processing.",
			"⌛ Processing..",
			"⌛ Processing...",
		},

		Encouragement: []string{
			"🔄 Still here, still working!",
			"💪 Haven't forgotten about you!",
			"🎯 Making progress, promise!",
			"✨ Almost there, hang tight!",
			"🚀 Full speed ahead!",
			"🌟 Your patience is appreciated!",
			"🔥 Working hard on this one!",
			"💫 Magic takes time!",
		},

		// EEHORIZON-specific tips and facts
		TipsAndFacts: []string{
			// Platform Features
			"💡 Pro tip: Use the ISO Selector in the top navigation bar to switch between ERCOT, PJM, MISO, and other markets!",
			"🗺️ Did you know? EEHorizon covers 8 major ISOs: ERCOT, PJM, MISO, CAISO, NYISO, ISO-NE, SPP, and WECC!",
			"📊 Fun fact: You can overlay multiple layers to see substations, resources, and constraints all at once!",
			"🎨 Tip: Use the Basemap Gallery to switch between Topographic, Dark Gray Canvas, and Streets views!",
			"📁 Did you know? You can upload your own KML, KMZ, CSV, GeoJSON, or Shapefiles using the Add Data feature!",
			"🔍 Pro tip: Use the Filter feature to narrow down data - it's mandatory for contour maps to appear!",
			"✏️ Fun fact: The Draw feature lets you create pins, lines, rectangles, polygons, and circles on the map!",
			"📤 Tip: You can export selected features to KML, CSV, or GeoJSON format using the Select feature!",
			"📍 Did you know? The Search bar lets you find elements by ISO name, layer, and column name!",
			"📏 Pro tip: Use the Measure tool to calculate area and distance in various units - just like Google Maps!",

			// Substations & Capacity
			"🔌 Pro tip: Each substation shows injection capacity AND $/MW estimates for interconnection costs!",
			"⚠️ Did you know? SSR (Sub-synchronous Resonance) risk levels are shown for each Point of Interconnection!",
			"📈 Tip: Lower N-X count means HIGHER SSR risk - keep an eye on those lighter areas on the contour map!",
			"🎯 Fun fact: Injection Capacity shows how much power can be transmitted without compromising grid reliability!",

			// LMP & Pricing
			"💰 Did you know? LMP Basis is the difference between a node's LMP and the zonal LMP!",
			"📊 Tip: Darker colors on LMP contour maps indicate higher Locational Marginal Pricing!",
			"⚡ Fun fact: Scarcity pricing can cause temporary distortions in LMP values during high demand!",
			"📈 Pro tip: Use LMP contour maps to identify congestion areas and optimal generation locations!",

			// Binding Constraints
			"🔗 Did you know? Shadow Price shows savings in production costs if a constraint is relaxed by 1 MW!",
			"📋 Tip: The Top 50 Binding Constraints are ranked by monthly shadow price ($/MWmonth)!",
			"🗺️ Fun fact: Rows highlighted in red in binding constraints represent unmapped constraints!",
			"⏰ Pro tip: Use the Month Filter to track how binding constraints change over time!",

			// Resources
			"🔋 Did you know? Planned Resources are projects with an Interconnection Agreement AND financial commitment!",
			"☀️ Tip: Operational Resources are categorized by fuel type with color-coded legends!",
			"📅 Fun fact: Large Load Data includes loads greater than 75 MW - think data centers and manufacturing plants!",

			// SROT & Battery Storage
			"🔋 Pro tip: SROT shows battery storage revenue estimates for 2024-2053 at the zonal level!",
			"💵 Did you know? SROT breaks down revenue into Energy Arbitrage (EA) and Ancillary Services (AS)!",
			"⚙️ Fun fact: SROT assumes 100 MW BESS with 2-hour duration and 80% round-trip efficiency!",
			"📊 Tip: Toggle the Decomposition option in SROT for a waterfall chart breakdown!",

			// Ancillary Services
			"⚡ Did you know? Ancillary Services help maintain grid stability by balancing supply and demand!",
			"🔄 Tip: ERCOT's Responsive Reserve Service provides standby reserves for grid disruptions!",
			"📈 Fun fact: Spinning Reserves must be online and available within 10 minutes!",
			"💡 Pro tip: Use AS data to estimate revenue and optimize market offers for power generators!",

			// Transmission
			"🔌 Did you know? Nationwide Transmission Lines connect power generation to load centers across the U.S.!",
			"📋 Tip: Planned Transmission Upgrades show project status, voltage, estimated cost, and expected service date!",
			"🗺️ Fun fact: Each ISO has unique terminology - ERCOT uses 'TPIT', PJM uses 'RTEP', MISO uses 'TEP'!",

			// Natural Resources
			"🌬️ Did you know? Wind turbines operate most efficiently within a specific 'rated wind speed' range!",
			"☀️ Tip: Solar energy production is directly correlated with irradiance - optimize panel orientation!",
			"⛽ Fun fact: EEHorizon shows coal mines, oil & gas wells, pipelines, and refining facilities!",

			// Regional Info
			"🗺️ Did you know? ERCOT operates independently from other U.S. interconnections!",
			"📍 Tip: PJM covers 13 states plus D.C. - from New Jersey to Illinois!",
			"🌐 Fun fact: WECC covers the entire Western U.S. including parts of Canada and Mexico!",
			"⚡ Pro tip: MISO spans from the Gulf of Mexico to the Canadian border!",

			// User Guide & Help
			"📖 Tip: Access the User Guide from the bottom left of the map interface for detailed instructions!",
			"ℹ️ Did you know? The Info feature shows study assumptions, data sources, and cost chart definitions!",
			"💾 Pro tip: Save your drawings using the Floppy Disk icon and reload them anytime with the Circular Arrow!",
			"🗑️ Tip: Click the Trash Can icon without any selection to delete ALL drawn elements at once!",

			// Export Limits
			"📤 Did you know? The maximum number of features you can export at once is 50!",
			"💡 Tip: Use the Cloud Icon with Down Arrow to download and save your drawings locally!",

			// Chart Features
			"📊 Pro tip: The Daily Average LMP Chart shows historical pricing data with color-coded peak types!",
			"📈 Tip: Use the Time Scale filter in SROT to adjust the time range by dragging the button!",
			"🔄 Did you know? The Reconductoring Cost Chart shows upgrade costs at various network levels!",

			// Selection Tools
			"🖱️ Tip: Use Rectangle, Circle, Lasso, Point, or Line selection tools for different selection needs!",
			"🔄 Pro tip: Toggle the Mode button to switch between Select and Deselect modes!",
			"📊 Did you know? The Information Bar shows the count of currently selected features!",

			// Draw Features
			"✏️ Tip: Use the Curved Line Tool by pressing right-click to start and releasing when done!",
			"📝 Pro tip: Prepare text using the Format button before placing it with the Text Tool!",
			"🎨 Did you know? The Format Tool lets you customize width, pattern, color, and opacity!",
			"📏 Tip: Enable the Measurement Tool before drawing to see area and distance values!",
		},

		ToolStart: map[string][]string{
			"get_layer_info": {
				"🗺️ Ah, layer intel requested! Let me dig into the details...",
				"📊 Time to unveil the secrets of this layer...",
				"🔬 Putting this layer under the microscope...",
				"🧐 Examining the layer specifications...",
				"📋 Pulling up the layer dossier...",
				"🔍 Zooming into layer details...",
				"📑 Opening the layer file cabinet...",
				"🎯 Targeting layer information...",
			},
			"get_all_available_layers": {
				"📋 Gathering the full layer catalog for you...",
				"🗂️ Let me fetch the complete layer menu!",
				"🌐 Scanning the entire layer universe...",
				"📚 Opening the grand layer library...",
				"🗃️ Pulling out the complete layer inventory...",
				"🔭 Surveying all available layers...",
				"📡 Scanning for all layer signals...",
				"🎪 Rolling out the full layer showcase...",
			},
			"locate_a_layer": {
				"📍 Layer location services activated!",
				"🗺️ Playing hide and seek with layers... I'm pretty good at this!",
				"🔎 Triangulating layer position... GPS eat your heart out!",
				"🧭 Compass pointing to your layer...",
				"🎯 Locking onto layer coordinates...",
				"🔍 Layer tracker engaged!",
				"📡 Pinging layer location...",
				"🌍 Mapping out where this layer lives...",
			},
			"get_user_guide_info": {
				"📖 Cracking open the user guide... I actually read these!",
				"📚 Consulting the sacred texts of EEHORIZON...",
				"🎓 Time for a quick knowledge transfer!",
				"📕 Flipping through the manual pages...",
				"🔖 Finding the relevant chapter for you...",
				"📜 Unrolling the scroll of wisdom...",
				"🧠 Accessing the knowledge base...",
				"💡 Let me enlighten you with the guide...",
			},
			"get_help_support": {
				"🆘 Help is on the way! Cape optional.",
				"🤝 Let me connect you with the right support...",
				"💡 Support beacon activated!",
				"🦸 Super support mode: ON!",
				"🎯 Targeting the help you need...",
				"📞 Dialing up support info...",
				"🔧 Fetching the troubleshooting toolkit...",
				"🌟 Your helpful guide is here!",
			},
			"get_layer_update_info": {
				"🆕 Checking for updates... unlike my phone, I do this promptly!",
				"📰 Fetching the latest layer news...",
				"🔄 Scanning for recent changes...",
				"📅 Checking the update calendar...",
				"🗞️ Hot off the press! Getting update info...",
				"🔔 Checking what's new in layer land...",
				"📊 Pulling the latest change log...",
				"⏰ Time-traveling to find recent updates...",
			},
		},

		ToolEnd: map[string][]string{
			"get_layer_info": {
				"✅ Got the layer scoop!",
				"📊 Layer details acquired!",
				"🎯 Found the info you need!",
				"📋 Layer intel gathered!",
				"🔍 Investigation complete!",
				"💎 Layer gems uncovered!",
				"📑 Dossier compiled!",
				"🏆 Layer mastery achieved!",
			},
			"get_all_available_layers": {
				"✅ Layer catalog ready!",
				"📋 Got the complete list!",
				"🗂️ All layers accounted for!",
				"📚 Full inventory acquired!",
				"🎪 The whole show is here!",
				"🌐 Global layer scan complete!",
				"📡 All signals received!",
				"🗃️ Cabinet fully opened!",
			},
			"locate_a_layer": {
				"📍 Layer located!",
				"🎯 Found it!",
				"✅ Target acquired!",
				"🧭 Location locked!",
				"🗺️ Mapped and marked!",
				"🔍 Discovery made!",
				"📡 Signal traced!",
				"🌍 Coordinates confirmed!",
			},
			"get_user_guide_info": {
				"📖 Found the relevant guide info!",
				"✅ Knowledge extracted!",
				"🎓 Got what you need!",
				"📚 Wisdom retrieved!",
				"🔖 Chapter found!",
				"💡 Enlightenment achieved!",
				"📕 Manual consulted!",
				"🧠 Info downloaded to my brain!",
			},
			"get_help_support": {
				"🤝 Support info ready!",
				"✅ Help is here!",
				"💡 Got your answer!",
				"🦸 Rescue complete!",
				"🎯 Help targeted!",
				"🔧 Toolkit prepared!",
				"📞 Connection made!",
				"🌟 Guidance obtained!",
			},
			"get_layer_update_info": {
				"🆕 Update info retrieved!",
				"✅ Got the latest changes!",
				"📰 News flash ready!",
				"📅 Schedule checked!",
				"🗞️ Headlines gathered!",
				"🔔 Notifications reviewed!",
				"📊 Changelog compiled!",
				"⏰ Timeline captured!",
			},
		},
	}
}

// Global instance for easy access
var wittyMessages = DefaultWittyMessages()

// GetRandomMessage returns a random message from a slice
func GetRandomMessage(messages []string) string {
	if len(messages) == 0 {
		return ""
	}
	return messages[rand.Intn(len(messages))]
}

// GetToolStartMessage returns a witty start message for a tool
func GetToolStartMessage(toolName string, fallback string) string {
	if messages, ok := wittyMessages.ToolStart[toolName]; ok && len(messages) > 0 {
		return GetRandomMessage(messages)
	}
	return fallback
}

// GetToolEndMessage returns a witty end message for a tool
func GetToolEndMessage(toolName string, fallback string) string {
	if messages, ok := wittyMessages.ToolEnd[toolName]; ok && len(messages) > 0 {
		return GetRandomMessage(messages)
	}
	return fallback
}

// GetThinkingMessage returns a random thinking message
func GetThinkingMessage() string {
	return GetRandomMessage(wittyMessages.Thinking)
}

// GetMultiStepMessage returns a random multi-step progress message
func GetMultiStepMessage() string {
	return GetRandomMessage(wittyMessages.MultiStep)
}

// GetFinalAnswerMessage returns a random final answer message
func GetFinalAnswerMessage() string {
	return GetRandomMessage(wittyMessages.FinalAnswer)
}

// GetErrorRecoveryMessage returns a random error recovery message
func GetErrorRecoveryMessage() string {
	return GetRandomMessage(wittyMessages.ErrorRecovery)
}

// GetLongWaitMessage returns a random long wait message
func GetLongWaitMessage() string {
	return GetRandomMessage(wittyMessages.LongWait)
}

// GetGreetingMessage returns a random greeting message
func GetGreetingMessage() string {
	return GetRandomMessage(wittyMessages.Greeting)
}

// GetTipOrFact returns a random EEHORIZON tip or fact
func GetTipOrFact() string {
	return GetRandomMessage(wittyMessages.TipsAndFacts)
}

// GetEncouragementMessage returns a random encouragement message
func GetEncouragementMessage() string {
	return GetRandomMessage(wittyMessages.Encouragement)
}

// GetDynamicWaitMessage returns a message based on elapsed time
// Mixes tips, encouragement, and long wait messages
func GetDynamicWaitMessage(elapsedSeconds int) string {
	// Mix of message types based on elapsed time
	if elapsedSeconds < 4 {
		return GetEncouragementMessage()
	} else if elapsedSeconds < 8 {
		// 50% chance of tip, 50% chance of encouragement
		if rand.Intn(2) == 0 {
			return GetTipOrFact()
		}
		return GetEncouragementMessage()
	} else if elapsedSeconds < 16 {
		// Tip or fact
		return GetTipOrFact()
	} else {
		// After 8 seconds, show tips and long wait messages
		switch rand.Intn(3) {
		case 0:
			return GetTipOrFact()
		case 1:
			return GetLongWaitMessage()
		default:
			return GetEncouragementMessage()
		}
	}
}

// ProgressTicker manages sending periodic progress messages
type ProgressTicker struct {
	ticker    *time.Ticker
	done      chan bool
	startTime time.Time
	isRunning bool
}

// NewProgressTicker creates a new progress ticker
func NewProgressTicker() *ProgressTicker {
	return &ProgressTicker{
		done: make(chan bool),
	}
}

// Start begins sending periodic progress messages
func (pt *ProgressTicker) Start(ctx context.Context, sendMessage func(msg types.StreamMessage) bool, interval time.Duration) {
	if pt.isRunning {
		return
	}

	pt.ticker = time.NewTicker(interval)
	pt.startTime = time.Now()
	pt.isRunning = true

	go func() {
		for {
			select {
			case <-pt.ticker.C:
				elapsed := int(time.Since(pt.startTime).Seconds())
				msg := GetDynamicWaitMessage(elapsed)
				sendMessage(types.StreamMessage{
					Type:    "info",
					Message: msg,
				})
			case <-pt.done:
				return
			case <-ctx.Done():
				return
			}
		}
	}()
}

// Stop stops the progress ticker
func (pt *ProgressTicker) Stop() {
	if pt.isRunning && pt.ticker != nil {
		pt.ticker.Stop()
		pt.done <- true
		pt.isRunning = false
	}
}

// AddToolMessages allows adding custom messages for new tools
func AddToolMessages(toolName string, startMessages, endMessages []string) {
	if len(startMessages) > 0 {
		wittyMessages.ToolStart[toolName] = startMessages
	}
	if len(endMessages) > 0 {
		wittyMessages.ToolEnd[toolName] = endMessages
	}
}

// AddThinkingMessages allows adding more thinking messages
func AddThinkingMessages(messages []string) {
	wittyMessages.Thinking = append(wittyMessages.Thinking, messages...)
}

// AddMultiStepMessages allows adding more multi-step messages
func AddMultiStepMessages(messages []string) {
	wittyMessages.MultiStep = append(wittyMessages.MultiStep, messages...)
}

// AddFinalAnswerMessages allows adding more final answer messages
func AddFinalAnswerMessages(messages []string) {
	wittyMessages.FinalAnswer = append(wittyMessages.FinalAnswer, messages...)
}

// AddTipsAndFacts allows adding more EEHORIZON tips and facts
func AddTipsAndFacts(messages []string) {
	wittyMessages.TipsAndFacts = append(wittyMessages.TipsAndFacts, messages...)
}
