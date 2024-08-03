package main

import (
	"errors"
	"fmt"
	"github.com/minhnghia2k3/greenlight/internal/data"
	"github.com/minhnghia2k3/greenlight/internal/validation"
	"net/http"
)

type Input struct {
	Title   *string       `json:"title"`
	Year    *int32        `json:"year"`
	Runtime *data.Runtime `json:"runtime"`
	Genres  []string      `json:"genres"` // Don't need to set to a pointer, bc slices already heave zero-values nil
}

type ListMovies struct {
	Data     data.Movie    `json:"data"`
	Metadata data.Metadata `json:"metadata"`
}

type MovieResponse struct {
	Movie data.Movie `json:"movie"`
}

// @Summary      Create movie
// @Description  handlers receives Input, validate it then create a new movie record
// @Tags         Movies
// @Accept 		 json
// @Produce      json
// @Security Bearer
// @Success      201  {object} MovieResponse
// @Failure      400  {object} Error
// @Failure      422  {object} Error
// @Failure      500  {object} Error
// @Router       /movies [post]
func (app *application) createMovieHandler(w http.ResponseWriter, r *http.Request) {
	var input Input

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err.Error())
		return
	}

	// ======== VALIDATING DATA ===========
	// Copy the values from the input struct to a new Movie struct
	movie := &data.Movie{
		Title:   *input.Title,
		Year:    *input.Year,
		Runtime: *input.Runtime,
		Genres:  input.Genres,
	}

	v := validation.New()

	if data.ValidateMovie(v, movie); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return

	} // Store data.
	err = app.models.Movies.Insert(movie)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// Sending HTTP response included location header to let client
	// know which URL they can find the newly-created
	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/movies/%d", movie.ID))

	err = app.writeJSON(w, http.StatusCreated, envelop{"movie": movie}, headers)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

// The listMoviesHandler GET list of movies, query by `input` struct
// It will get the URL query, query integer parameters, and response list of movies to client.
// @Summary      List movies
// @Description  show list movies, page = 1, page_size=10 by default.
// @Param page query int false "page"
// @Param page_size query int false "page_size"
// @Param title query string false "title"
// @Param genres query string false "genres"
// @Param sort query string false "sort"
// @Security Bearer
// @Tags         Movies
// @Accept 		 json
// @Produce      json
// @Success      200  {object} ListMovies
// @Failure      500  {object} Error
// @Router       /movies [get]
func (app *application) listMoviesHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Title  string
		Genres []string
		data.Filters
	}

	v := validation.New()
	qs := r.URL.Query()

	// Read query parameters
	input.Title = app.readString(qs, "title", "")
	input.Genres = app.readCSV(qs, "genres", []string{})
	input.Page = app.readInt(qs, "page", 1, v)
	input.PageSize = app.readInt(qs, "page_size", 10, v)
	input.Sort = app.readString(qs, "sort", "id")
	input.SortSafeList = []string{"id", "title", "year", "runtime", "-id", "-title", "-year", "-runtime"}
	if data.ValidateFilters(v, input.Filters); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	// Get list movies
	movies, metadata, err := app.models.Movies.GetAll(input.Title, input.Genres, input.Filters)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	err = app.writeJSON(w, http.StatusOK, envelop{"metadata": metadata, "movies": movies}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

// showMovieHandler handler
// @Summary      Get movie by id
// @Description  get movie by provided movie id
// @Param id path int true "id"
// @Tags         Movies
// @Accept 		 json
// @Produce      json
// @Security Bearer
// @Success      200  {object} ListMovies
// @Failure      404  {object} Error
// @Failure      500  {object} Error
// @Router       /movies/{id} [get]
func (app *application) showMovieHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIdParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	movie, err := app.models.Movies.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}

		return
	}

	err = app.writeJSON(w, http.StatusOK, envelop{"movie": movie}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

// updateMovieHandler which query to get a movie by parameter id and update with input variables.
// @Summary      Update movie
// @Description  update an existing movie record
// @Param id path int true "id"
// @Param input query Input true "update movie input"
// @Security Bearer
// @Tags         Movies
// @Accept 		 json
// @Produce      json
// @Success      200  {object} MovieResponse
// @Failure      400  {object} Error
// @Failure      404  {object} Error
// @Failure      422  {object} Error
// @Failure      500  {object} Error
// @Router       /movies [patch]
func (app *application) updateMovieHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIdParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	// Fetch the corresponding movie record
	movie, err := app.models.Movies.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	// Read input data then validate
	var input Input

	err = app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err.Error())
		return
	}

	if input.Title != nil {
		movie.Title = *input.Title
	}

	if input.Year != nil {
		movie.Year = *input.Year
	}

	if input.Runtime != nil {
		movie.Runtime = *input.Runtime
	}

	if input.Genres != nil {
		movie.Genres = input.Genres
	}

	v := validation.New()

	if data.ValidateMovie(v, movie); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	// Update to store the updated movie record in our database.
	err = app.models.Movies.Update(movie)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrConflictEdit):
			app.editConflictResponse(w, r)
			return
		default:
			app.serverErrorResponse(w, r, err)
			return
		}
	}

	err = app.writeJSON(w, http.StatusOK, envelop{"movies": movie}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

// @Summary      Delete movie
// @Description  delete a movie record
// @Param id path int true "id"
// @Param   input      body Input true  "update movie input"
// @Security Bearer
// @Tags         Movies
// @Accept 		 json
// @Produce      json
// @Success      200  {object} MovieResponse
// @Failure      404  {object} Error
// @Failure      500  {object} Error
// @Router       /movies [delete]
func (app *application) deleteMovieHandler(w http.ResponseWriter, r *http.Request) {
	id, err := app.readIdParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	err = app.models.Movies.Delete(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelop{"message": "movie successfully deleted"}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
